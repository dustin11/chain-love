package d_util

import (
	"chain-love/domain/active"
	"chain-love/domain/auth"
	"chain-love/domain/ds"
	"chain-love/domain/sys"
	"chain-love/pkg/setting"
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

var tables = []interface{}{
	&sys.User{},
	&auth.AuthNonce{},
	&auth.RefreshToken{},

	&ds.Book{},
	&ds.Image{},
	&ds.Note{},

	&active.Like{},
}

func InitTable(db *gorm.DB) {
	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			err := db.Migrator().CreateTable(table)
			if err != nil {
				log.Fatal("create table err: ", err.Error())
				break
			}
			log.Printf("create table %v success.", table)
		}
		// auto‑migrate – 保证结构体中的新字段自动加入已有表
		if err := db.AutoMigrate(table); err != nil {
			log.Printf("automigrate %T failed: %v", table, err)
		}
	}

	// ensure ds_book auto_increment starts at 10000 (MySQL). Safe no-op if DB/dialect differs.
	if err := EnsureTableAutoIncrement(db, "ds_book", 10000); err != nil {
		log.Printf("ensure auto_increment ds_book failed: %v", err)
	}
	if err := EnsureTableAutoIncrement(db, "ds_image", 1000); err != nil {
		log.Printf("ensure auto_increment ds_image failed: %v", err)
	}
	if err := EnsureTableAutoIncrement(db, "ds_note", 1000); err != nil {
		log.Printf("ensure auto_increment ds_note failed: %v", err)
	}
}

// EnsureTableAutoIncrement ensures the AUTO_INCREMENT for a MySQL table is at least start.
// It's a no-op on DBs that don't support information_schema AUTO_INCREMENT.
func EnsureTableAutoIncrement(db *gorm.DB, table string, start int64) error {
	var ai sql.NullInt64
	row := db.Raw(`
        SELECT AUTO_INCREMENT
        FROM information_schema.TABLES
        WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?
    `, table).Row()

	if err := row.Scan(&ai); err != nil {
		return fmt.Errorf("read auto_increment failed: %w", err)
	}

	if !ai.Valid || ai.Int64 < start {
		// 有些 MySQL 语句不能使用占位符，这里直接把数字拼进去（表名用反引号转义）
		sqlStr := fmt.Sprintf("ALTER TABLE `%s` AUTO_INCREMENT = %d", table, start)
		if err := db.Exec(sqlStr).Error; err != nil {
			return fmt.Errorf("set auto_increment failed: %w", err)
		}
	}
	return nil
}

// EnsureDatabaseExists connects to MySQL without selecting a database and
// creates the database if it does not exist. Caller must ensure the
// configured MySQL user has CREATE DATABASE privilege.
func EnsureDatabaseExists(name string) error {
	dsnNoDB := fmt.Sprintf("%s:%s@tcp(%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		setting.Config.Database.User,
		setting.Config.Database.Password,
		setting.Config.Database.Host,
	)

	sqlDB, err := sql.Open("mysql", dsnNoDB)
	if err != nil {
		return fmt.Errorf("open mysql (no db) failed: %w", err)
	}
	defer sqlDB.Close()

	_, err = sqlDB.Exec(fmt.Sprintf(
		"CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci",
		name,
	))
	if err != nil {
		return fmt.Errorf("create database %s failed: %w", name, err)
	}
	return nil
}

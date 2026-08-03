package d_util

import (
	"database/sql"
	"fmt"
	"log"
	"senspace/domain/active"
	"senspace/domain/auth"
	"senspace/domain/dev"
	"senspace/domain/ds"
	"senspace/domain/factory"
	"senspace/domain/planet/road"
	"senspace/domain/planet/terrain"
	"senspace/domain/sys"
	"senspace/domain/task"
	"senspace/pkg/setting"

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
	&ds.PluginAsset{},
	&ds.PluginAssetBinding{},
	&ds.PluginInstanceState{},
	&ds.PluginInstanceDraft{},
	&ds.PluginComment{},
	&ds.PluginShare{},

	&active.Like{},

	&dev.Plugin{},
}

func InitTable(db *gorm.DB) {
	allTables := append([]interface{}{}, tables...)
	allTables = append(allTables, factory.Tables()...)
	allTables = append(allTables, road.Tables()...)
	allTables = append(allTables, terrain.Tables()...)
	allTables = append(allTables, task.Tables()...)
	for _, table := range allTables {
		if !db.Migrator().HasTable(table) {
			err := db.Migrator().CreateTable(table)
			if err != nil {
				log.Fatal("create table err: ", err.Error())
				break
			}
			log.Printf("create table %v success.", table)
		}
		if _, ok := table.(*ds.PluginInstanceState); ok {
			if err := migratePluginInstanceStateColumns(db); err != nil {
				log.Printf("migrate plugin instance state columns failed: %v", err)
			}
		}
		if _, ok := table.(*ds.PluginShare); ok {
			if err := dropUnusedPluginShareColumns(db); err != nil {
				log.Printf("drop unused plugin share columns failed: %v", err)
			}
		}
		// auto‑migrate – 保证结构体中的新字段自动加入已有表
		if err := db.AutoMigrate(table); err != nil {
			log.Printf("automigrate %T failed: %v", table, err)
		}
		if err := migrateAuditTimeColumns(db, table); err != nil {
			log.Printf("migrate audit time columns %T failed: %v", table, err)
		}
	}
	if err := factory.DropUnusedNFTInventoryPoolColumns(db); err != nil {
		log.Printf("drop unused nft inventory pool columns failed: %v", err)
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

	factory.SeedBuiltinReleases(db)
}

// dropUnusedPluginShareColumns 删除不再参与分享运行时的旧字段。
func dropUnusedPluginShareColumns(db *gorm.DB) error {
	table := &ds.PluginShare{}
	if db.Migrator().HasColumn(table, "background_key") {
		return db.Migrator().DropColumn(table, "background_key")
	}
	return nil
}

// migratePluginInstanceStateColumns 将旧资源状态和位姿字段迁移到新的状态字段。
func migratePluginInstanceStateColumns(db *gorm.DB) error {
	table := &ds.PluginInstanceState{}
	if db.Migrator().HasColumn(table, "state_json") &&
		!db.Migrator().HasColumn(table, "resource_state_json") {
		if err := db.Migrator().RenameColumn(table, "state_json", "resource_state_json"); err != nil {
			return err
		}
	}
	if db.Migrator().HasColumn(table, "pose_json") &&
		!db.Migrator().HasColumn(table, "state_json") {
		if err := db.Migrator().RenameColumn(table, "pose_json", "state_json"); err != nil {
			return err
		}
	}
	return nil
}

// migrateAuditTimeColumns 将旧审计字段 created_on/updated_on 迁移到 CreatedAt/UpdatedAt 对应列。
func migrateAuditTimeColumns(db *gorm.DB, table interface{}) error {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(table); err != nil {
		return err
	}
	tableName := stmt.Schema.Table
	if tableName == "" {
		return nil
	}

	if err := migrateAuditTimeColumn(db, table, tableName, "created_on", "created_at"); err != nil {
		return err
	}
	return migrateAuditTimeColumn(db, table, tableName, "updated_on", "updated_at")
}

func migrateAuditTimeColumn(db *gorm.DB, table interface{}, tableName string, oldColumn string, newColumn string) error {
	if !db.Migrator().HasColumn(table, oldColumn) || !db.Migrator().HasColumn(table, newColumn) {
		return nil
	}
	copySQL := fmt.Sprintf(
		"UPDATE `%s` SET `%s` = `%s` WHERE `%s` IS NOT NULL AND `%s` IS NULL",
		tableName,
		newColumn,
		oldColumn,
		oldColumn,
		newColumn,
	)
	if err := db.Exec(copySQL).Error; err != nil {
		return err
	}
	return db.Migrator().DropColumn(table, oldColumn)
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

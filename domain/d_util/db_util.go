package d_util

import (
	"chain-love/domain/ds"
	"chain-love/domain/rent"
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
	&ds.House{},
	&ds.Room{},
	&rent.Publish{},
	&rent.Rent{},
	&rent.RentHistory{},
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
	}
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

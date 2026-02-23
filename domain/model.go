package domain

import (
	"chain-love/pkg/app/contextx"
	"chain-love/pkg/logging"
	"chain-love/pkg/setting"
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Db *gorm.DB

type CreatInfo struct {
	CreatedOn time.Time `json:"-" gorm:"autoCreateTime"`
	CreatedBy uint64    `json:"-"`
}

type UpdateInfo struct {
	UpdatedOn time.Time `json:"-" gorm:"autoUpdateTime"`
	UpdatedBy uint64    `json:"-"`
}

func Setup() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		setting.Config.Database.User,
		setting.Config.Database.Password,
		setting.Config.Database.Host,
		setting.Config.Database.Name)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		logging.Error(err)
		return
	}
	Db = db

	// configure connection pool
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(10)
	}
}

// 需要从api 传入ctx.Gin.Request.Context() 太麻烦，不用
// GORM v2 hooks: 从 tx.Statement.Context 读取 user_id 自动填充 CreatedBy/ModifiedBy
func (m *CreatInfo) BeforeCreate(tx *gorm.DB) (err error) {
	if v := tx.Statement.Context.Value(contextx.CtxUserID); v != nil {
		if id, ok := v.(uint64); ok {
			m.CreatedBy = id
		}
	}
	return nil
}

func (m *UpdateInfo) BeforeUpdate(tx *gorm.DB) (err error) {
	if v := tx.Statement.Context.Value(contextx.CtxUserID); v != nil {
		if id, ok := v.(uint64); ok {
			m.UpdatedBy = id
		}
	}
	return nil
}

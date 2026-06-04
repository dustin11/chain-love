package domain

import (
	"fmt"
	"senspace/pkg/app/contextx"
	"senspace/pkg/logging"
	"senspace/pkg/setting"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Db *gorm.DB

const (
	initialConnectMaxAttempts = 30
	initialConnectRetryDelay  = 2 * time.Second
)

type CreatInfo struct {
	CreatedAt time.Time `json:"-" gorm:"autoCreateTime"`
	CreatedBy uint64    `json:"-"`
}

type UpdateInfo struct {
	UpdatedAt time.Time `json:"-" gorm:"autoUpdateTime"`
	UpdatedBy uint64    `json:"-"`
}

func Setup() error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		setting.Config.Database.User,
		setting.Config.Database.Password,
		setting.Config.Database.Host,
		setting.Config.Database.Name)

	open := func() (*gorm.DB, error) {
		return gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
	}

	db, err := openWithRetry(open, initialConnectMaxAttempts, initialConnectRetryDelay)
	if err != nil {
		logging.Error("mysql open failed", err)
		return err
	}
	Db = db

	configurePool(db)

	// 启动后台循环，定时 ping 数据库并在连接断开时自动重连
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := Db.Exec("SELECT 1").Error; err != nil {
				logging.Warn("mysql ping failed, trying reconnect", err)
				reconnect(open)
			}
		}
	}()

	return nil
}

// configurePool 将通用的连接池设置提取出来，方便重连时复用
func configurePool(db *gorm.DB) {
	if sqlDB, err := db.DB(); err == nil {
		sqlDB.SetMaxOpenConns(10)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(time.Hour)
	}
}

// reconnect 持续尝试重新建立数据库连接
func reconnect(open func() (*gorm.DB, error)) {
	for {
		db, err := open()
		if err == nil {
			Db = db
			configurePool(db)
			logging.Info("mysql reconnected successfully")
			return
		}
		logging.Error("mysql reconnect failed", err)
		time.Sleep(5 * time.Second)
	}
}

func openWithRetry(open func() (*gorm.DB, error), maxAttempts int, delay time.Duration) (*gorm.DB, error) {
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	if delay <= 0 {
		delay = time.Second
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		db, err := open()
		if err == nil {
			return db, nil
		}
		lastErr = err
		if attempt == maxAttempts {
			break
		}
		logging.Warn(fmt.Sprintf("mysql open attempt %d/%d failed, retrying in %s", attempt, maxAttempts, delay), err)
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("open mysql after %d attempts: %w", maxAttempts, lastErr)
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

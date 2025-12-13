package consts

import (
	"os"
	"time"
)

const (
	TOKEN_PREFIX  = "Bearer "
	AUTHORIZATION = "Authorization"
	ACCESS_TOKEN  = "access_token"
	REFRESH_TOKEN = "refresh_token"

	ADMIN        = 1
	MENU_PAGE_ID = 4
)

// 全局可配置的 token 时长（以 time.Duration 表示）
var (
	// 访问令牌有效期，默认 1 小时
	AccessTokenTTL = time.Hour

	// 刷新令牌有效期，默认 30 天
	RefreshTokenTTL = 30 * 24 * time.Hour
)

func Getenv() string {
	env := os.Getenv("SPIDER_ENV")
	if env == "" {
		env = "dev"
	}
	return env
}

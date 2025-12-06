package consts

import "os"

const (
	TOKEN_PREFIX  = "Bearer "
	AUTHORIZATION = "Authorization"

	ADMIN        = 1
	MENU_PAGE_ID = 4
)

func Getenv() string {
	env := os.Getenv("SPIDER_ENV")
	if env == "" {
		env = "dev"
	}
	return env
}

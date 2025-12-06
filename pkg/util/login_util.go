package util

import (
	"chain-love/pkg/app"
	"chain-love/pkg/e"
	"chain-love/pkg/setting/consts"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetToken(c *gin.Context) string {
	token := c.Query(consts.AUTHORIZATION)
	if token == "" {
		token = c.Request.Header.Get(consts.AUTHORIZATION)
	}
	if token == "" {
		token = c.Request.Header.Get("token")
	}
	if token == "" {
		panic(e.ParameterError("no token"))
	}
	if !strings.HasPrefix(token, consts.TOKEN_PREFIX) {
		token, _ = url.QueryUnescape(token)
		if !strings.HasPrefix(token, consts.TOKEN_PREFIX) {
			panic(e.ParameterError("error token str"))
		}
	}
	return strings.Fields(token)[1]
}

func GetTokenUser(c *gin.Context) app.JwtUser {
	token := GetToken(c)
	user, err := app.ParseToken(token)
	e.PanicIfErr(err)

	return user
}

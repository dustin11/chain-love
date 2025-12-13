package util

import (
	"chain-love/pkg/app/security"
	"chain-love/pkg/e"
	"chain-love/pkg/setting/consts"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

func GetBearerTokenFromHeader(c *gin.Context) string {
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

func GetTokenUser(c *gin.Context) security.JwtUser {
	token := GetBearerTokenFromHeader(c)
	user, err := security.ParseToken(token)
	e.PanicIfErr(err)

	return user
}

package middleware

import (
	"chain-love/domain"
	"chain-love/pkg/app"
	"chain-love/pkg/app/contextx"
	"chain-love/pkg/util"
	"context"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	passed := []string{"/api/v1/home/register", "/api/v1/home/login", "/api/v1/home/logout", "/api/v1/basic/houseType"}

	return func(ctx *gin.Context) {
		if ok, _ := util.Contain(ctx.Request.URL.Path, passed); ok {
			ctx.Next()
			return
		}
		if ChkCookie(ctx) {
			ctx.Next()
			return
		}
		if ok, user := ChkToken(ctx); ok {
			// 把整个 user 放到 gin.Context，供 handler 使用
			ctx.Set("user", user)

			// 把 user_id 放到 request.Context，供 GORM 钩子通过 tx.Statement.Context 读取
			reqCtx := context.WithValue(ctx.Request.Context(), contextx.CtxUserID, user.Id)
			ctx.Request = ctx.Request.WithContext(reqCtx)

			// 可选：把带请求 context 的 DB 放到 gin.Context，业务层可直接取用
			ctx.Set("db", domain.Db.WithContext(reqCtx))

			ctx.Next()
			return
		}

		// 未授权
		ctx.AbortWithStatus(401)
	}
}

func ChkCookie(ctx *gin.Context) bool {
	_, e := ctx.Request.Cookie("user_cookie")
	if e == nil {
		return true
	} else {
		return false
	}
}

func ChkToken(c *gin.Context) (bool, *app.JwtUser) {
	user := util.GetTokenUser(c)
	return true, &user
}

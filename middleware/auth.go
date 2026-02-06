package middleware

import (
	"chain-love/domain"
	"chain-love/pkg/app/contextx"
	"chain-love/pkg/app/security"
	"chain-love/pkg/setting/consts"
	"context"
	"strings"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	passed := []string{"/api/v1/home/register", "/api/v1/home/login",
		"/api/v1/home/logout"} //, "/api/v1/user/addr/"

	return func(ctx *gin.Context) {
		for _, p := range passed {
			if ctx.Request.URL.Path == p || strings.HasPrefix(ctx.Request.URL.Path, p+"/") || strings.HasPrefix(ctx.Request.URL.Path, p) {
				ctx.Next()
				return
			}
		}
		if ok, user := ChkToken(ctx); ok {
			applyUserToContext(ctx, user)
			ctx.Next()
			return
		}
		// 未授权
		ctx.AbortWithStatus(401)
	}
}

func ChkCookie(ctx *gin.Context) bool {
	_, e := ctx.Request.Cookie(consts.ACCESS_TOKEN)
	if e == nil {
		return true
	} else {
		return false
	}
}

func ChkToken(c *gin.Context) (bool, *security.JwtUser) {
	// 1. 先尝试从 HttpOnly cookie 获取（优先）
	token := ""
	if ck, err := c.Request.Cookie(consts.ACCESS_TOKEN); err == nil {
		token = ck.Value
	}
	// 2. 否则尝试 Authorization header（Bearer ...）
	// if token == "" {
	// 	token = util.GetBearerTokenFromHeader(c)
	// }
	if token == "" {
		return false, nil
	}

	// 3. 解析并验证 JWT（签名、exp/nbf/iss/aud/sub 等）
	user, err := security.ParseToken(token)
	if err != nil {
		// 解析失败：无效或过期
		return false, nil
	}

	// 4. 检查撤销/黑名单（建议使用 token jti 或 token hash）
	//    AuthTokenStore 可用 Redis 缓存以减轻 DB 压力
	// if AuthTokenStore.IsRevoked(user.JTI) {
	// 	return false, nil
	// }

	// 5. 可选：校验绑定信息（IP/UA）
	// if !checkBinding(user.JTI, c.ClientIP(), c.Request.UserAgent()) { return false, nil }

	// 6. 一切通过：把 user 返回
	return true, &user
}

func applyUserToContext(ctx *gin.Context, user *security.JwtUser) {
	// 把整个 user 放到 gin.Context，供 handler 使用
	ctx.Set("user", user)

	// 把 user_id 放到 request.Context，供 GORM 钩子通过 tx.Statement.Context 读取
	reqCtx := context.WithValue(ctx.Request.Context(), contextx.CtxUserID, user.Id)
	ctx.Request = ctx.Request.WithContext(reqCtx)

	// 可选：把带请求 context 的 DB 放到 gin.Context，业务层可直接取用
	ctx.Set("db", domain.Db.WithContext(reqCtx))
}

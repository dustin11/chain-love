package auth_api

import (
	"chain-love/pkg/app"
	"chain-love/pkg/e"
	"chain-love/pkg/logging"
	"chain-love/pkg/setting/consts"
	"chain-love/service/auth_service"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// @Summary 获取登录 Nonce
// @Tags 认证
// @Param		addr			path		string	true	"用户addr"
// @Success 200 {object} e.Error
// @Router /api/v1/auth/nonce/{addr} [get]
func GetNonce(c *gin.Context) {
	addr := c.Param("addr")
	// 参数校验改为统一抛出参数错误
	e.PanicIfParameterError(addr == "", "地址不能为空")
	e.PanicIfParameterError(!strings.HasPrefix(addr, "0x") || len(addr) != 42, "无效的钱包地址")

	nonce := auth_service.GenerateNonce(strings.ToLower(addr))

	app.Response(c, e.SuccessData(map[string]string{"nonce": nonce}))
}

type LoginReq struct {
	Message   string `json:"message" binding:"required"`
	Signature string `json:"signature" binding:"required"`
}

// @Summary SIWE 登录
// @Tags 认证
// @Param data body LoginReq true "登录数据"
// @Success 200 {object} e.Error
// @Router /api/v1/auth/verify [post]
func Verify(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		e.PanicIfParameterError(err != nil, "参数错误")
	}

	// auth_service 在内部遇到问题会直接 panic
	accessToken, refreshToken, user := auth_service.VerifyAndLogin(
		req.Message,
		req.Signature,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	// 设置 Cookie
	setAuthCookies(c, accessToken, refreshToken)

	app.Response(c, e.SuccessData(map[string]interface{}{
		// "access_token":  accessToken,
		// "refresh_token": refreshToken,
		"user": user.ToJwtUser(),
	}))
}

// @Summary 刷新 Token
// @Tags 认证
// @Success 200 {object} e.Error
// @Router /api/v1/auth/refresh [post]
func Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(consts.REFRESH_TOKEN)
	e.PanicIfParameterError(err != nil || refreshToken == "", "缺少刷新令牌")

	// auth_service 内部会在失败时 panic
	newAccess, newRefresh := auth_service.RefreshToken(refreshToken, c.ClientIP())

	setAuthCookies(c, newAccess, newRefresh)
	app.Response(c, e.Success)
}

// @Summary 登出
// @Tags 认证
// @Success 200 {object} e.Error
// @Router /api/v1/auth/logout [post]
func Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie(consts.REFRESH_TOKEN)
	if refreshToken != "" {
		// 错误不中断登出流程
		if err := auth_service.Logout(refreshToken); err != nil {
			logging.Error("Logout error:", err)
		}
	}
	clearAuthCookies(c)
	app.Response(c, e.Success)
}

// 设置 Cookie
// func setAuthCookies(c *gin.Context, access, refresh string) {
// 	// Access Token: 短期，HttpOnly 时间以秒为单位
// 	c.SetCookie(consts.ACCESS_TOKEN, access, 60*10, "/", "", false, true)
// 	// Refresh Token: 长期，HttpOnly, Path 限定在 refresh 接口更安全（可选）
// 	c.SetCookie(consts.REFRESH_TOKEN, refresh, 3600*24*30, "/", "", false, true)
// }

func setAuthCookies(ctx *gin.Context, access, refresh string) {
	// 使用 consts 中统一的过期配置，MaxAge 以秒为单位
	accessMaxAge := int(consts.AccessTokenTTL.Seconds())
	refreshMaxAge := int(consts.RefreshTokenTTL.Seconds())

	// Access token cookie
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     consts.ACCESS_TOKEN,
		Value:    access,
		Path:     "/", // 若需更严格限制，改为 "/api/v1/auth" 或其它路径
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   accessMaxAge,
	})

	// Refresh token cookie（可限制 Path 到 refresh 接口以降低暴露面）
	http.SetCookie(ctx.Writer, &http.Cookie{
		Name:     consts.REFRESH_TOKEN,
		Value:    refresh,
		Path:     "/api/v1/auth", // 建议限定到刷新相关路径
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   refreshMaxAge,
	})
}

// 清除 Cookie
func clearAuthCookies(c *gin.Context) {
	c.SetCookie(consts.ACCESS_TOKEN, "", -1, "/", "", false, true)
	c.SetCookie(consts.REFRESH_TOKEN, "", -1, "/api/v1/auth", "", false, true)
}

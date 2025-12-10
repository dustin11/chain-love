package auth_api

import (
	"chain-love/pkg/app"
	"chain-love/pkg/e"
	"chain-love/service/auth_service"
	"strings"

	"github.com/gin-gonic/gin"
)

// @Summary 获取登录 Nonce
// @Tags 认证
// @Param address query string true "钱包地址"
// @Success 200 {object} app.Response
// @Router /api/v1/auth/nonce [get]
func GetNonce(c *gin.Context) {
	addr := c.Query("address")
	// 参数校验改为统一抛出参数错误
	e.PanicIfParameterError(addr == "", "地址不能为空")
	e.PanicIfParameterError(!strings.HasPrefix(addr, "0x") || len(addr) != 42, "无效的钱包地址")

	nonce := auth_service.GenerateNonce(strings.ToLower(addr))

	app.Response(c, e.SuccessData(nonce))
}

type LoginReq struct {
	Message   string `json:"message" binding:"required"`
	Signature string `json:"signature" binding:"required"`
}

// @Summary SIWE 登录
// @Tags 认证
// @Param data body LoginReq true "登录数据"
// @Success 200 {object} app.Response
// @Router /api/v1/auth/login [post]
func Login(c *gin.Context) {
	var req LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		e.PanicIfParameterError(err != nil, "参数错误")
	}

	// auth_service 在内部遇到问题会直接 panic
	accessToken, refreshToken := auth_service.VerifyAndLogin(
		req.Message,
		req.Signature,
		c.ClientIP(),
		c.Request.UserAgent(),
	)

	// 设置 Cookie
	setAuthCookies(c, accessToken, refreshToken)

	app.Response(c, e.SuccessData(map[string]string{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}))
}

// @Summary 刷新 Token
// @Tags 认证
// @Success 200 {object} app.Response
// @Router /api/v1/auth/refresh [post]
func Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	e.PanicIfParameterError(err != nil || refreshToken == "", "缺少刷新令牌")

	// auth_service 内部会在失败时 panic
	newAccess, newRefresh := auth_service.RefreshSession(refreshToken, c.ClientIP())

	setAuthCookies(c, newAccess, newRefresh)
	app.Response(c, e.SuccessData(map[string]string{
		"access_token": newAccess,
	}))
}

// @Summary 登出
// @Tags 认证
// @Success 200 {object} app.Response
// @Router /api/v1/auth/logout [post]
func Logout(c *gin.Context) {
	refreshToken, _ := c.Cookie("refresh_token")
	if refreshToken != "" {
		// auth_service 会在必要时 panic
		auth_service.Logout(refreshToken)
	}
	clearAuthCookies(c)
	app.Response(c, e.SuccessData("已登出"))
}

// 辅助：设置 Cookie
func setAuthCookies(c *gin.Context, access, refresh string) {
	// Access Token: 短期，HttpOnly
	c.SetCookie("access_token", access, 3600, "/", "", false, true)
	// Refresh Token: 长期，HttpOnly, Path 限定在 refresh 接口更安全（可选）
	c.SetCookie("refresh_token", refresh, 3600*24*30, "/", "", false, true)
}

// 辅助：清除 Cookie
func clearAuthCookies(c *gin.Context) {
	c.SetCookie("access_token", "", -1, "/", "", false, true)
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)
}

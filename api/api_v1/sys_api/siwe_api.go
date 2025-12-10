package sys_api

// import (
// 	"chain-love/pkg/app"
// 	"chain-love/pkg/e"
// 	"chain-love/service/user_service"

// 	"github.com/gin-gonic/gin"
// 	"github.com/spruceid/siwe-go"
// )

// // @Summary 获取SIWE Nonce
// // @Tags SIWE登录
// // @Success 200 {object} app.Response
// // @Router /api/v1/home/siwe/nonce [get]
// func SiweNonce(c *gin.Context) {
// 	// 生成随机 Nonce
// 	nonce := siwe.GenerateNonce()

// 	// 将 Nonce 存入 HttpOnly Cookie，有效期 5 分钟，防止 XSS 窃取
// 	c.SetCookie("siwe_nonce", nonce, 300, "/", "", false, true)

// 	app.Response(c, e.SuccessData(nonce))
// }

// type SiweLoginReq struct {
// 	Message   string `json:"message"`
// 	Signature string `json:"signature"`
// }

// // @Summary SIWE 登录
// // @Tags SIWE登录
// // @Param data body SiweLoginReq true "SIWE数据"
// // @Success 200 {object} app.Response
// // @Router /api/v1/home/siwe/login [post]
// func SiweLogin(c *gin.Context) {
// 	var req SiweLoginReq
// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		app.Response(c, e.Error(e.INVALID_PARAMS, "参数错误"))
// 		return
// 	}

// 	// 1. 从 Cookie 获取并验证 Nonce
// 	nonce, err := c.Cookie("siwe_nonce")
// 	if err != nil || nonce == "" {
// 		app.Response(c, e.Error(e.ERROR_AUTH_CHECK_TOKEN_FAIL, "Nonce无效或已过期，请重新获取"))
// 		return
// 	}

// 	// 2. 解析 SIWE 消息
// 	msg, err := siwe.ParseMessage(req.Message)
// 	if err != nil {
// 		app.Response(c, e.Error(e.INVALID_PARAMS, "消息格式错误"))
// 		return
// 	}

// 	// 3. 校验消息中的 Nonce 是否与 Cookie 一致（防重放）
// 	if msg.GetNonce() != nonce {
// 		app.Response(c, e.Error(e.ERROR_AUTH_CHECK_TOKEN_FAIL, "Nonce不匹配"))
// 		return
// 	}

// 	// 4. 验证签名 (siwe-go 会自动校验格式、过期时间等)
// 	if _, err := msg.Verify(req.Signature, nil, &nonce, nil); err != nil {
// 		app.Response(c, e.Error(e.ERROR_AUTH_CHECK_TOKEN_FAIL, "签名验证失败"))
// 		return
// 	}

// 	// 5. 验证通过，获取以太坊地址并执行登录逻辑
// 	addr := msg.GetAddress().String()
// 	token, err := user_service.LoginByAddr(addr)
// 	if err != nil {
// 		app.Response(c, e.Error(e.ERROR, "登录失败"))
// 		return
// 	}

// 	// 登录成功，清除 Nonce Cookie
// 	c.SetCookie("siwe_nonce", "", -1, "/", "", false, true)

// 	app.Response(c, e.SuccessData(map[string]string{
// 		"token": token,
// 		"addr":  addr,
// 	}))
// }

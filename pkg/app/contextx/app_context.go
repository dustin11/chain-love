package contextx

import (
	"chain-love/pkg/app"
	"chain-love/pkg/app/security"
	"chain-love/pkg/e"
	"chain-love/pkg/logging"

	// "chain-love/pkg/util"

	"github.com/gin-gonic/gin"
)

type AppContext struct {
	Gin  *gin.Context
	User *security.JwtUser
}

func (c *AppContext) setUp() {
	// 优先使用 middleware 已经注入的 user（避免重复解析）
	if u, ok := c.Gin.Get("user"); ok {
		if ju, ok := u.(*security.JwtUser); ok {
			c.User = ju
			return
		}
	}
	// 若 middleware 未注入（例如某些公开路由）
	logging.Warn("AppContext setup no user found")
	// // 回退：若 middleware 未注入（例如某些公开路由），再尝试从 Header 解析
	// token := util.GetBearerTokenFromHeader(c.Gin)
	// if token == "" {
	// 	return
	// }
	// if user, err := security.ParseToken(token); err == nil {
	// 	c.User = &user
	// } else {
	// 	// 不在这里 panic，避免影响请求流程；middleware 应负责统一鉴权/错误转换
	// 	log.Printf("ParseToken, fail to parseToken: %v", err)
	// }
}

func (c *AppContext) Response(e *e.Error) {
	app.Response(c.Gin, e)
}

type AppHandleFunc func(c *AppContext)

func WithAppContext(appHandle AppHandleFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		//fmt.Println(<- chan struct{}(nil))

		// 可以在gin.Context中设置key-value
		c.Set("trace", "假设这是一个调用链追踪sdk")

		// 全局超时控制
		//timeoutCtx, cancelFunc := context.WithTimeout(c, 5 * time.Second)
		//defer cancelFunc()

		// ZDM上下文
		//appCtx := AppContext{Context: timeoutCtx, Gin: c}
		appCtx := AppContext{Gin: c}
		appCtx.setUp()
		// 回调接口
		appHandle(&appCtx)
	}
}

package contextx

import (
	"chain-love/pkg/app"
	"chain-love/pkg/e"
	"chain-love/pkg/util"
	"log"

	"github.com/gin-gonic/gin"
)

type AppContext struct {
	//context.Context
	Gin  *gin.Context
	User *app.JwtUser
}

func (c *AppContext) setUp() {
	token := util.GetToken(c.Gin)
	if user, err := app.ParseToken(token); err == nil {
		c.User = &user
	} else {
		log.Printf("ParseToken, fail to parseToken: %v", err)
		panic(err)
	}
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

package routers

import (
	"chain-love/asset"
	"chain-love/middleware"
	"chain-love/pkg/e"
	"chain-love/pkg/setting"
	"chain-love/pkg/util"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alecthomas/template"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-contrib/cors"
)

func SetupRouter() *gin.Engine {
	router := gin.New()

	// 明确列出受信任的代理地址或 CIDR（不要用 "*" 或 nil 去信任所有代理）
	// 开发环境常用 localhost；生产环境请填写你的负载均衡/反向代理的 IP 或网段
	if err := router.SetTrustedProxies([]string{
		"127.0.0.1",        // localhost (dev)
		"10.0.0.0/8",       // 示例：内部网段或云内网段
		// "203.0.113.5",    // 或具体代理 IP
	}); err != nil {
		log.Println("SetTrustedProxies failed:", err)
	}

	// 使用显式的 CORS 配置：允许凭证且必须指定具体 origin（不能用 "*"）
	corsCfg := cors.Config{
		AllowOrigins:     setting.Config.App.AllowedCORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,           // 必须启用，浏览器才能发送/接收带凭证的跨域 Cookie
		MaxAge:           12 * time.Hour, // 预检请求的缓存时间
	}

	router.Use(middleware.Logger(), gin.Recovery(), middleware.ErrHandler(), cors.New(corsCfg)) //middleware.Cors()
	router.NoMethod(e.HandleNotFound)
	router.NoRoute(e.HandleNotFound)

	tmpPath := "asset/statics/templates/*"
	if gin.Mode() == gin.TestMode {
		tmpPath = "./../asset/statics/templates/*"
	}
	router.LoadHTMLGlob(tmpPath)
	router.Static("asset/statics", ".assert/statics")
	router.StaticFile("/favicon.ico", "./favicon.ico")
	router.StaticFS("/avatar", http.Dir(util.RootPath()+"avatar/"))
	//swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	SetupApiV1Router(router)

	return router
}

func htmlAssets() {
	//r := multitemplate.New()
	bytes, err := asset.Asset("temp/oldindex.html")
	if err != nil {
		log.Println(err)
		return
	}
	t, err := template.New("index").Parse(string(bytes))
	log.Println(t, err)
	//r.Add("index", t)
	//router.HTMLRender = r
}

func helloGinAndMethod(ctx *gin.Context) {
	ctx.String(http.StatusOK, "hello gin "+strings.ToLower(ctx.Request.Method)+" method")
}

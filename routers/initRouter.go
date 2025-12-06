package routers

import (
	"chain-love/asset"
	"chain-love/middleware"
	"chain-love/pkg/e"
	"chain-love/pkg/util"
	"log"
	"net/http"
	"strings"

	"github.com/alecthomas/template"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func SetupRouter() *gin.Engine {
	//router := gin.Default()
	router := gin.New()
	//router.Use(middleware.Context(), middleware.Logger(), gin.Recovery(), middleware.ErrHandler(), middleware.Cors())
	router.Use(middleware.Logger(), gin.Recovery(), middleware.ErrHandler(), middleware.Cors())
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
	//url := ginSwagger.URL("http://localhost:9000/swagger/doc.json")
	//router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	//index := router.Group("/")
	//{
	//	// 添加 Get 请求路由
	//	index.GET("/", handler.Index)
	//
	//	index.POST("/", helloGinAndMethod)
	//
	//	index.PUT("/", helloGinAndMethod)
	//
	//	index.DELETE("/", helloGinAndMethod)
	//
	//}
	//
	//userRouter := router.Group("/user", middleware.Auth())
	//{
	//	userRouter.POST("/register", user.Register)
	//	userRouter.POST("/login", user.Login)
	//	userRouter.GET("/profile", user.Profile)
	//	userRouter.POST("/update", user.UpdateProfile)
	//}

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

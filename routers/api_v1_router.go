package routers

import (
	"chain-love/api/api_v1/auth_api"
	"chain-love/api/api_v1/ds_api"
	"chain-love/api/api_v1/sys_api"
	"chain-love/middleware"
	context "chain-love/pkg/app/contextx"

	"github.com/gin-gonic/gin"
)

func SetupApiV1Router(router *gin.Engine) {
	// 公开访问的路由
	// 认证相关
	authRouter := router.Group("/api/v1/auth")
	{
		authRouter.GET("/nonce/:addr", auth_api.GetNonce)
		authRouter.POST("/verify", auth_api.Verify)
		authRouter.POST("/refresh", auth_api.Refresh)
		authRouter.POST("/logout", auth_api.Logout)
	}

	// 需要认证的 API 路由
	apiRouter := router.Group("/api/v1", middleware.Auth())

	homeRouter := apiRouter.Group("/home")
	{
		// homeRouter.POST("/register", sys_api.Register)
		homeRouter.POST("/login", sys_api.Login)
		homeRouter.POST("/logout", sys_api.Logout)

		homeRouter.GET("/info", context.WithAppContext(sys_api.UserInfo))
	}
	userRouter := apiRouter.Group("/user")
	{
		userRouter.GET("/me", context.WithAppContext(ds_api.Me))
		userRouter.GET("/page", ds_api.UserGetPage)
		userRouter.GET("/info/:id", ds_api.UserGetById)
		userRouter.GET("/addr/:addr", ds_api.UserGetByAddr)
		userRouter.POST("/save", context.WithAppContext(ds_api.UserSave))
		userRouter.POST("/del/:id", ds_api.UserDel)
	}

	bookRouter := apiRouter.Group("/book")
	{
		bookRouter.GET("/page", ds_api.BookGetPage)
		bookRouter.GET("/info/:id", ds_api.BookGetById)
		bookRouter.POST("/save", context.WithAppContext(ds_api.BookSave))
		bookRouter.POST("/del/:id", ds_api.BookDel)
	}
	imageRouter := apiRouter.Group("/image")
	{
		imageRouter.POST("/save", context.WithAppContext(ds_api.ImageSave))
	}
	//基础数据
	basicRouter := apiRouter.Group("/basic")
	{
		basicRouter.GET("/decorationType", ds_api.GetDecorationType)
		basicRouter.GET("/furnitureType", ds_api.GetFurnitureType)
		basicRouter.GET("/houseType", ds_api.GetHouseType)
	}

}

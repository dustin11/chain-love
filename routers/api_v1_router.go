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
	apiRouter := router.Group("/api/v1", middleware.Auth())
	// 认证相关路由 (公开)
	authRouter := router.Group("/api/v1/auth")
	{
		authRouter.GET("/nonce", auth_api.GetNonce)
		authRouter.POST("/login", auth_api.Login)
		authRouter.POST("/refresh", auth_api.Refresh)
		authRouter.POST("/logout", auth_api.Logout)
	}
	homeRouter := apiRouter.Group("/home")
	{
		// homeRouter.POST("/register", sys_api.Register)
		homeRouter.POST("/login", sys_api.Login)
		homeRouter.POST("/logout", sys_api.Logout)

		homeRouter.GET("/info", context.WithAppContext(sys_api.UserInfo))
	}
	userRouter := apiRouter.Group("/user")
	{
		userRouter.GET("/page", sys_api.UserGetPage)
		userRouter.GET("/info/:id", sys_api.UserGetById)
		userRouter.GET("/addr/:addr", sys_api.UserGetByAddr)
		userRouter.POST("/save", context.WithAppContext(sys_api.UserSave))
		userRouter.POST("/del/:id", sys_api.UserDel)
	}

	//基础数据
	basicRouter := apiRouter.Group("/basic")
	{
		basicRouter.GET("/decorationType", ds_api.GetDecorationType)
		basicRouter.GET("/furnitureType", ds_api.GetFurnitureType)
		basicRouter.GET("/houseType", ds_api.GetHouseType)
	}

}

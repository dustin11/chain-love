package routers

import (
	"chain-love/api/api_v1/ds_api"
	"chain-love/api/api_v1/sys_api"
	"chain-love/middleware"
	context "chain-love/pkg/app/contextx"

	"github.com/gin-gonic/gin"
)

func SetupApiV1Router(router *gin.Engine) {
	apiRouter := router.Group("/api/v1", middleware.Auth())
	//apiRouter := router.Group("/api/v1")
	homeRouter := apiRouter.Group("/home")
	{
		homeRouter.POST("/register", sys_api.Register)
		homeRouter.POST("/login", sys_api.Login)
		homeRouter.POST("/logout", sys_api.Logout)
		homeRouter.GET("/info", context.WithAppContext(sys_api.UserInfo))
	}
	userRouter := apiRouter.Group("/user")
	{
		userRouter.GET("/page", sys_api.UserGetPage)
		userRouter.GET("/info/:id", sys_api.UserGetInfo)
		userRouter.POST("/save", context.WithAppContext(sys_api.UserSave))
		userRouter.POST("/del/:id", sys_api.UserDel)
	}
	menuRouter := apiRouter.Group("/menu")
	{
		menuRouter.GET("/one/:id", sys_api.GetMenuInfo)
		menuRouter.GET("/page", context.WithAppContext(sys_api.GetPage))
		menuRouter.GET("/listPerms", context.WithAppContext(sys_api.GetListWithPerms))
		menuRouter.POST("/save", sys_api.SaveMenu)
		menuRouter.POST("/del/:id", sys_api.DeleteMenu)
	}
	roleRouter := apiRouter.Group("/role")
	{
		roleRouter.GET("/info/:id", sys_api.GetRoleInfo)
		roleRouter.GET("/page", sys_api.GetRolePage)
		roleRouter.POST("/save", context.WithAppContext(sys_api.SaveRole))
		roleRouter.POST("/del/:id", sys_api.DeleteRole)
	}

	//基础数据
	basicRouter := apiRouter.Group("/basic")
	{
		basicRouter.GET("/decorationType", ds_api.GetDecorationType)
		basicRouter.GET("/furnitureType", ds_api.GetFurnitureType)
		basicRouter.GET("/houseType", ds_api.GetHouseType)
		basicRouter.GET("/orientationType", ds_api.GetOrientationType)
		basicRouter.GET("/rentInclude", ds_api.GetRentInclude)
		basicRouter.GET("/rentPayType", ds_api.GetRentPayType)
		basicRouter.GET("/rentState", ds_api.GetRentState)
		basicRouter.GET("/rentType", ds_api.GetRentType)
	}
	//房源
	houseRouter := apiRouter.Group("/house")
	{
		houseRouter.GET("/info/:id", ds_api.GetHouseInfo)
		houseRouter.GET("/page", ds_api.GetHousePage)
		houseRouter.POST("/save", context.WithAppContext(ds_api.SaveHouse))
		houseRouter.POST("/del/:id", ds_api.DeleteHouse)
	}

}

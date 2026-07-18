package routers

import (
	"senspace/api/api_v1/active_api"
	"senspace/api/api_v1/auth_api"
	"senspace/api/api_v1/dev_api"
	"senspace/api/api_v1/ds_api"
	"senspace/api/api_v1/factory_api"
	"senspace/api/api_v1/plugin_comment_api"
	"senspace/api/api_v1/plugin_share_api"
	"senspace/api/api_v1/sys_api"
	"senspace/api/api_v1/task_api"
	"senspace/middleware"
	context "senspace/pkg/app/contextx"

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
	imgRouter := router.Group("/api/v1/image")
	{
		imgRouter.GET("/list", ds_api.ImageList)
	}

	// 互动
	activePublicRouter := router.Group("/api/v1/active")
	{
		activePublicRouter.POST("/like/info", active_api.LikeGetSummary)
	}

	// 笔记（公开列表，鉴权保存和删除）
	noteRouter := router.Group("/api/v1/note")
	{
		noteRouter.GET("/list", ds_api.NoteList)
	}

	// 通过星球ID查询书籍ID列表（无需鉴权）
	bookPub := router.Group("/api/v1/book")
	{
		bookPub.GET("/ids/planet/:planetId", ds_api.BookGetIdByPlanetId)
	}
	factoryPublicRouter := router.Group("/api/v1/factory")
	{
		factoryPublicRouter.GET("/market", factory_api.GetPublicMarketList)
	}
	// 开发插件源码按当前约定公开可读，动态插件分享与开发工作台复用同一加载接口。
	// 修改、删除等工作区操作仍保留在下方鉴权路由中。
	pluginPublicRouter := router.Group("/api/v1/plugin")
	{
		pluginPublicRouter.GET("/tree/:pluginId", dev_api.GetPluginTree)
	}
	pluginCommentPublicRouter := router.Group(
		"/api/v1/plugin-comments",
		middleware.OptionalAuth(),
	)
	{
		pluginCommentPublicRouter.POST(
			"/list",
			context.WithAppContext(plugin_comment_api.ListComments),
		)
	}
	pluginSharePublicRouter := router.Group(
		"/api/v1/planet-moments",
		middleware.OptionalAuth(),
	)
	{
		pluginSharePublicRouter.GET(
			"/:shareToken/bootstrap",
			context.WithAppContext(plugin_share_api.GetBootstrap),
		)
		pluginSharePublicRouter.GET(
			"/:shareToken/snapshot",
			context.WithAppContext(plugin_share_api.GetBootstrap),
		)
		pluginSharePublicRouter.GET(
			"/:shareToken/resources/:resourceAlias",
			context.WithAppContext(plugin_share_api.GetResource),
		)
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
	// 图片
	imageRouter := apiRouter.Group("/image")
	{
		imageRouter.POST("/save", context.WithAppContext(ds_api.ImageSave))
		imageRouter.DELETE("/del/:id", context.WithAppContext(ds_api.ImageDel))
	}
	// 插件实例资源
	pluginAssetRouter := apiRouter.Group("/plugin-assets")
	{
		pluginAssetRouter.POST("/fact/:factAssetId/upload", context.WithAppContext(ds_api.PluginAssetUploadFact))
		pluginAssetRouter.POST("/fact/:factAssetId/import-image", context.WithAppContext(ds_api.PluginAssetImportImageFact))
		pluginAssetRouter.POST("/fact/:factAssetId/commit", context.WithAppContext(ds_api.PluginAssetCommitFact))
		pluginAssetRouter.PUT("/fact/:factAssetId/state", context.WithAppContext(ds_api.PluginAssetSaveStateFact))
		pluginAssetRouter.DELETE("/fact/:factAssetId/assets/:assetId", context.WithAppContext(ds_api.PluginAssetDeleteFact))
		pluginAssetRouter.POST("/fact/:factAssetId/snapshot/rebuild", context.WithAppContext(ds_api.PluginAssetRebuildSnapshotFact))

		pluginAssetRouter.POST("/dev/:pluginId/:version/upload", context.WithAppContext(ds_api.PluginAssetUploadDev))
		pluginAssetRouter.POST("/dev/:pluginId/:version/import-image", context.WithAppContext(ds_api.PluginAssetImportImageDev))
		pluginAssetRouter.POST("/dev/:pluginId/:version/commit", context.WithAppContext(ds_api.PluginAssetCommitDev))
		pluginAssetRouter.PUT("/dev/:pluginId/:version/state", context.WithAppContext(ds_api.PluginAssetSaveStateDev))
		pluginAssetRouter.DELETE("/dev/:pluginId/:version/assets/:assetId", context.WithAppContext(ds_api.PluginAssetDeleteDev))
		pluginAssetRouter.POST("/dev/:pluginId/:version/snapshot/rebuild", context.WithAppContext(ds_api.PluginAssetRebuildSnapshotDev))

		pluginAssetRouter.POST("/draft/:releaseId/:draftId/commit", context.WithAppContext(ds_api.PluginAssetCommitDraft))
		pluginAssetRouter.GET("/draft/:releaseId/active", context.WithAppContext(ds_api.PluginAssetGetActiveDraft))

		pluginAssetRouter.POST("/:factAssetId/upload", context.WithAppContext(ds_api.PluginAssetUpload))
		pluginAssetRouter.POST("/:factAssetId/import-image", context.WithAppContext(ds_api.PluginAssetImportImage))
		pluginAssetRouter.PUT("/:factAssetId/state", context.WithAppContext(ds_api.PluginAssetSaveState))
		pluginAssetRouter.DELETE("/:factAssetId/assets/:assetId", context.WithAppContext(ds_api.PluginAssetDelete))
		pluginAssetRouter.POST("/:factAssetId/snapshot/rebuild", context.WithAppContext(ds_api.PluginAssetRebuildSnapshot))
	}
	// 互动-需鉴权
	activeRouter := apiRouter.Group("/active")
	{
		activeRouter.POST("/like/add", context.WithAppContext(active_api.LikeAdd))
		activeRouter.POST("/like/del", context.WithAppContext(active_api.LikeDel))
	}
	// 笔记
	noteAuthRouter := apiRouter.Group("/note")
	{
		noteAuthRouter.POST("/save", context.WithAppContext(ds_api.NoteSave))
		noteAuthRouter.POST("/del", context.WithAppContext(ds_api.NoteDel))
	}
	// 插件工作区写操作需要鉴权；源码树读取在公开路由中注册。
	pluginRouter := apiRouter.Group("/plugin")
	{
		pluginRouter.GET("/list", dev_api.GetPluginList)
		pluginRouter.GET("/versions/:pluginId", dev_api.GetPluginVersions)
		pluginRouter.POST("/file/upload", dev_api.UploadFile)
		pluginRouter.POST("/folder/add", dev_api.AddFolder)
		pluginRouter.POST("/rename", dev_api.Rename)
		pluginRouter.POST("/delete", dev_api.Delete)
		pluginRouter.POST("/deletePlugin", context.WithAppContext(dev_api.DeletePlugin))
		pluginRouter.POST("/save", context.WithAppContext(dev_api.SavePlugin))
	}
	// 插件评论
	pluginCommentRouter := apiRouter.Group("/plugin-comments")
	{
		pluginCommentRouter.POST("/create", context.WithAppContext(plugin_comment_api.CreateComment))
		pluginCommentRouter.POST("/like", context.WithAppContext(plugin_comment_api.LikeComment))
		pluginCommentRouter.POST("/cleanup", context.WithAppContext(plugin_comment_api.CleanupComments))
	}
	pluginShareRouter := apiRouter.Group("/planet-moments")
	{
		pluginShareRouter.GET("/my", context.WithAppContext(plugin_share_api.ListMyShares))
		pluginShareRouter.GET("/my/:momentId/bootstrap", context.WithAppContext(plugin_share_api.GetOwnedBootstrap))
		pluginShareRouter.DELETE("/my/:shareId", context.WithAppContext(plugin_share_api.DeleteManagedShare))
		pluginShareRouter.POST("", context.WithAppContext(plugin_share_api.CreateShare))
		pluginShareRouter.DELETE("/:shareToken", context.WithAppContext(plugin_share_api.DeleteShare))
	}
	//基础数据
	basicRouter := apiRouter.Group("/basic")
	{
		basicRouter.GET("/decorationType", ds_api.GetDecorationType)
		basicRouter.GET("/furnitureType", ds_api.GetFurnitureType)
		basicRouter.GET("/houseType", ds_api.GetHouseType)
	}
	factoryRouter := apiRouter.Group("/factory")
	{
		factoryRouter.POST("/publish", context.WithAppContext(factory_api.PublishPlugin))
		factoryRouter.POST("/plugins/:pluginId/freeze-current", context.WithAppContext(factory_api.FreezeCurrentPluginReleaseAssets))
		factoryRouter.POST("/plugins/:pluginId/clear-freeze-current", context.WithAppContext(factory_api.ClearCurrentPluginReleaseAssetsFreeze))
		factoryRouter.POST("/plugins/:pluginId/generate-asset-data", context.WithAppContext(factory_api.GenerateReleaseAssetData))
		factoryRouter.GET("/releases/my", context.WithAppContext(factory_api.GetMyReleases))
		factoryRouter.GET("/releases/:id", context.WithAppContext(factory_api.GetReleaseDetail))
		factoryRouter.POST("/releases/:id/mint", context.WithAppContext(factory_api.MintReleaseAsset))
		factoryRouter.POST("/releases/:id/clear-dev", context.WithAppContext(factory_api.ClearReleaseDev))
		factoryRouter.POST("/releases/:id/clear", context.WithAppContext(factory_api.ClearRelease))
		factoryRouter.PATCH("/releases/:id", context.WithAppContext(factory_api.UpdateRelease))
		factoryRouter.PATCH("/releases/:id/price", context.WithAppContext(factory_api.UpdateReleasePrice))
		factoryRouter.PATCH("/releases/:id/status", context.WithAppContext(factory_api.UpdateReleaseStatus))
		factoryRouter.GET("/ownership/my", context.WithAppContext(factory_api.GetMyOwnerships))
	}
	taskRouter := apiRouter.Group("/task")
	{
		taskRouter.GET("/list", context.WithAppContext(task_api.ListTasks))
		taskRouter.POST("/purge-all", context.WithAppContext(task_api.PurgeAllTasks))
		taskRouter.POST("/:id/retry", context.WithAppContext(task_api.RetryTask))
		taskRouter.DELETE("/:id", context.WithAppContext(task_api.DeleteTask))
	}

}

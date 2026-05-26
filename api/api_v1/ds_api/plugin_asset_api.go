package ds_api

import (
	"encoding/json"
	"strconv"

	"senspace/pkg/app"
	"senspace/pkg/app/contextx"
	"senspace/pkg/e"
	"senspace/service/ds_service"
)

func bindPluginAssetStateInput(ctx *contextx.AppContext) ds_service.PluginInstanceStateInput {
	var req ds_service.PluginInstanceStateInput
	err := ctx.Gin.ShouldBindJSON(&req)
	e.PanicParameterError(err)
	for index, binding := range req.Bindings {
		if binding.AssetId == 0 {
			e.PanicIfParameterError(true, "资源ID不能为空")
		}
		if len(binding.Config) > 0 && !json.Valid(binding.Config) {
			e.PanicIfParameterError(true, "资源配置JSON无效")
		}
		req.Bindings[index] = binding
	}
	return req
}

// @Summary 上传 fact 插件实例资源
// @Tags 插件资源
// @Param factAssetId path string true "插件资产实例ID"
// @Param collectionKey formData string false "资源集合键"
// @Param file formData file true "资源文件"
// @Security ApiKeyAuth
// @Router /api/v1/plugin-assets/fact/{factAssetId}/upload [post]
func PluginAssetUploadFact(ctx *contextx.AppContext) {
	factAssetId, err := strconv.ParseInt(ctx.Gin.Param("factAssetId"), 10, 64)
	e.PanicParameterError(err)
	file, err := ctx.Gin.FormFile("file")
	e.PanicParameterErrorTipMsg(err, "file missing")
	scope, err := ds_service.ResolveFactPluginAssetScope(*ctx.User, factAssetId)
	e.PanicServerErr(err)
	result, err := ds_service.UploadPluginAssetInScope(*ctx.User, scope, ctx.Gin.PostForm("collectionKey"), file)
	e.PanicServerErr(err)
	app.Response(ctx.Gin, e.SuccessData(result))
}

// @Summary 上传 dev 插件实例资源
// @Tags 插件资源
// @Param pluginId path string true "插件ID"
// @Param version path string true "插件版本"
// @Param collectionKey formData string false "资源集合键"
// @Param file formData file true "资源文件"
// @Security ApiKeyAuth
// @Router /api/v1/plugin-assets/dev/{pluginId}/{version}/upload [post]
func PluginAssetUploadDev(ctx *contextx.AppContext) {
	pluginId := ctx.Gin.Param("pluginId")
	version := ctx.Gin.Param("version")
	file, err := ctx.Gin.FormFile("file")
	e.PanicParameterErrorTipMsg(err, "file missing")
	scope, err := ds_service.ResolveDevPluginAssetScope(*ctx.User, pluginId, version)
	e.PanicServerErr(err)
	result, err := ds_service.UploadPluginAssetInScope(*ctx.User, scope, ctx.Gin.PostForm("collectionKey"), file)
	e.PanicServerErr(err)
	app.Response(ctx.Gin, e.SuccessData(result))
}

// @Summary 保存 fact 插件实例状态
// @Tags 插件资源
// @Param factAssetId path string true "插件资产实例ID"
// @Param body body ds_service.PluginInstanceStateInput true "实例状态"
// @Security ApiKeyAuth
// @Router /api/v1/plugin-assets/fact/{factAssetId}/state [put]
func PluginAssetSaveStateFact(ctx *contextx.AppContext) {
	factAssetId, err := strconv.ParseInt(ctx.Gin.Param("factAssetId"), 10, 64)
	e.PanicParameterError(err)
	scope, err := ds_service.ResolveFactPluginAssetScope(*ctx.User, factAssetId)
	e.PanicServerErr(err)
	snapshot, err := ds_service.SavePluginInstanceStateInScope(*ctx.User, scope, bindPluginAssetStateInput(ctx))
	e.PanicServerErr(err)
	app.Response(ctx.Gin, e.SuccessData(snapshot))
}

// @Summary 保存 dev 插件实例状态
// @Tags 插件资源
// @Param pluginId path string true "插件ID"
// @Param version path string true "插件版本"
// @Param body body ds_service.PluginInstanceStateInput true "实例状态"
// @Security ApiKeyAuth
// @Router /api/v1/plugin-assets/dev/{pluginId}/{version}/state [put]
func PluginAssetSaveStateDev(ctx *contextx.AppContext) {
	scope, err := ds_service.ResolveDevPluginAssetScope(*ctx.User, ctx.Gin.Param("pluginId"), ctx.Gin.Param("version"))
	e.PanicServerErr(err)
	snapshot, err := ds_service.SavePluginInstanceStateInScope(*ctx.User, scope, bindPluginAssetStateInput(ctx))
	e.PanicServerErr(err)
	app.Response(ctx.Gin, e.SuccessData(snapshot))
}

// @Summary 删除 fact 插件实例资源
// @Tags 插件资源
// @Param factAssetId path string true "插件资产实例ID"
// @Param assetId path string true "资源ID"
// @Security ApiKeyAuth
// @Router /api/v1/plugin-assets/fact/{factAssetId}/assets/{assetId} [delete]
func PluginAssetDeleteFact(ctx *contextx.AppContext) {
	factAssetId, err := strconv.ParseInt(ctx.Gin.Param("factAssetId"), 10, 64)
	e.PanicParameterError(err)
	assetId, err := strconv.ParseUint(ctx.Gin.Param("assetId"), 10, 64)
	e.PanicParameterError(err)
	scope, err := ds_service.ResolveFactPluginAssetScope(*ctx.User, factAssetId)
	e.PanicServerErr(err)
	snapshot, err := ds_service.DeletePluginAssetInScope(*ctx.User, scope, assetId)
	e.PanicServerErr(err)
	app.Response(ctx.Gin, e.SuccessData(snapshot))
}

// @Summary 删除 dev 插件实例资源
// @Tags 插件资源
// @Param pluginId path string true "插件ID"
// @Param version path string true "插件版本"
// @Param assetId path string true "资源ID"
// @Security ApiKeyAuth
// @Router /api/v1/plugin-assets/dev/{pluginId}/{version}/assets/{assetId} [delete]
func PluginAssetDeleteDev(ctx *contextx.AppContext) {
	assetId, err := strconv.ParseUint(ctx.Gin.Param("assetId"), 10, 64)
	e.PanicParameterError(err)
	scope, err := ds_service.ResolveDevPluginAssetScope(*ctx.User, ctx.Gin.Param("pluginId"), ctx.Gin.Param("version"))
	e.PanicServerErr(err)
	snapshot, err := ds_service.DeletePluginAssetInScope(*ctx.User, scope, assetId)
	e.PanicServerErr(err)
	app.Response(ctx.Gin, e.SuccessData(snapshot))
}

// @Summary 重建 fact 插件实例资源快照
// @Tags 插件资源
// @Param factAssetId path string true "插件资产实例ID"
// @Security ApiKeyAuth
// @Router /api/v1/plugin-assets/fact/{factAssetId}/snapshot/rebuild [post]
func PluginAssetRebuildSnapshotFact(ctx *contextx.AppContext) {
	factAssetId, err := strconv.ParseInt(ctx.Gin.Param("factAssetId"), 10, 64)
	e.PanicParameterError(err)
	scope, err := ds_service.ResolveFactPluginAssetScope(*ctx.User, factAssetId)
	e.PanicServerErr(err)
	snapshot, err := ds_service.RebuildPluginAssetSnapshot(scope)
	e.PanicServerErr(err)
	app.Response(ctx.Gin, e.SuccessData(snapshot))
}

// @Summary 重建 dev 插件实例资源快照
// @Tags 插件资源
// @Param pluginId path string true "插件ID"
// @Param version path string true "插件版本"
// @Security ApiKeyAuth
// @Router /api/v1/plugin-assets/dev/{pluginId}/{version}/snapshot/rebuild [post]
func PluginAssetRebuildSnapshotDev(ctx *contextx.AppContext) {
	scope, err := ds_service.ResolveDevPluginAssetScope(*ctx.User, ctx.Gin.Param("pluginId"), ctx.Gin.Param("version"))
	e.PanicServerErr(err)
	snapshot, err := ds_service.RebuildPluginAssetSnapshot(scope)
	e.PanicServerErr(err)
	app.Response(ctx.Gin, e.SuccessData(snapshot))
}

// 兼容旧 fact 路由。
func PluginAssetUpload(ctx *contextx.AppContext) {
	PluginAssetUploadFact(ctx)
}

// 兼容旧 fact 路由。
func PluginAssetSaveState(ctx *contextx.AppContext) {
	PluginAssetSaveStateFact(ctx)
}

// 兼容旧 fact 路由。
func PluginAssetDelete(ctx *contextx.AppContext) {
	PluginAssetDeleteFact(ctx)
}

// 兼容旧 fact 路由。
func PluginAssetRebuildSnapshot(ctx *contextx.AppContext) {
	PluginAssetRebuildSnapshotFact(ctx)
}

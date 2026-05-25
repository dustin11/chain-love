package ds_api

import (
	"encoding/json"
	"strconv"

	"senspace/pkg/app"
	"senspace/pkg/app/contextx"
	"senspace/pkg/e"
	"senspace/service/ds_service"
)

// @Summary 上传插件实例资源
// @Tags 插件资源
// @Param factAssetId path string true "插件资产实例ID"
// @Param collectionKey formData string false "资源集合键"
// @Param file formData file true "资源文件"
// @Security ApiKeyAuth
// @Router /api/v1/plugin-assets/{factAssetId}/upload [post]
func PluginAssetUpload(ctx *contextx.AppContext) {
	factAssetId, err := strconv.ParseInt(ctx.Gin.Param("factAssetId"), 10, 64)
	e.PanicParameterError(err)
	file, err := ctx.Gin.FormFile("file")
	e.PanicParameterErrorTipMsg(err, "file missing")
	result, err := ds_service.UploadPluginAsset(
		*ctx.User,
		factAssetId,
		ctx.Gin.PostForm("collectionKey"),
		file,
	)
	e.PanicServerErr(err)
	app.Response(ctx.Gin, e.SuccessData(result))
}

// @Summary 保存插件实例状态
// @Tags 插件资源
// @Param factAssetId path string true "插件资产实例ID"
// @Param body body ds_service.PluginInstanceStateInput true "实例状态"
// @Security ApiKeyAuth
// @Router /api/v1/plugin-assets/{factAssetId}/state [put]
func PluginAssetSaveState(ctx *contextx.AppContext) {
	factAssetId, err := strconv.ParseInt(ctx.Gin.Param("factAssetId"), 10, 64)
	e.PanicParameterError(err)
	var req ds_service.PluginInstanceStateInput
	err = ctx.Gin.ShouldBindJSON(&req)
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
	snapshot, err := ds_service.SavePluginInstanceState(*ctx.User, factAssetId, req)
	e.PanicServerErr(err)
	app.Response(ctx.Gin, e.SuccessData(snapshot))
}

// @Summary 删除插件实例资源
// @Tags 插件资源
// @Param factAssetId path string true "插件资产实例ID"
// @Param assetId path string true "资源ID"
// @Security ApiKeyAuth
// @Router /api/v1/plugin-assets/{factAssetId}/assets/{assetId} [delete]
func PluginAssetDelete(ctx *contextx.AppContext) {
	factAssetId, err := strconv.ParseInt(ctx.Gin.Param("factAssetId"), 10, 64)
	e.PanicParameterError(err)
	assetId, err := strconv.ParseUint(ctx.Gin.Param("assetId"), 10, 64)
	e.PanicParameterError(err)
	snapshot, err := ds_service.DeletePluginAsset(*ctx.User, factAssetId, assetId)
	e.PanicServerErr(err)
	app.Response(ctx.Gin, e.SuccessData(snapshot))
}

// @Summary 重建插件实例资源快照
// @Tags 插件资源
// @Param factAssetId path string true "插件资产实例ID"
// @Security ApiKeyAuth
// @Router /api/v1/plugin-assets/{factAssetId}/snapshot/rebuild [post]
func PluginAssetRebuildSnapshot(ctx *contextx.AppContext) {
	factAssetId, err := strconv.ParseInt(ctx.Gin.Param("factAssetId"), 10, 64)
	e.PanicParameterError(err)
	owner, err := ds_service.ResolvePluginAssetOwner(*ctx.User, factAssetId)
	e.PanicServerErr(err)
	snapshot, err := ds_service.RebuildPluginAssetSnapshot(owner.OwnerKey, factAssetId)
	e.PanicServerErr(err)
	app.Response(ctx.Gin, e.SuccessData(snapshot))
}

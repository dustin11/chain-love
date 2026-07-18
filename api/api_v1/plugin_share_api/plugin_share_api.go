package plugin_share_api

import (
	"strconv"
	"strings"

	"senspace/pkg/app"
	"senspace/pkg/app/contextx"
	"senspace/pkg/bizerr"
	"senspace/pkg/e"
	"senspace/service/plugin_share_service"
)

// CreateShare 创建单插件分享快照。
func CreateShare(ctx *contextx.AppContext) {
	e.PanicIfParameterError(ctx.User == nil, "请先登录")
	var input plugin_share_service.CreateInput
	e.PanicParameterError(ctx.Gin.ShouldBindJSON(&input))
	result, err := plugin_share_service.CreateShare(*ctx.User, input)
	bizerr.PanicHTTP(err)
	app.Response(ctx.Gin, e.SuccessData(result))
}

// GetBootstrap 返回分享页最小公开启动数据。
func GetBootstrap(ctx *contextx.AppContext) {
	result, err := plugin_share_service.GetBootstrap(ctx.Gin.Param("shareToken"), ctx.User)
	bizerr.PanicHTTP(err)
	ctx.Gin.Header("Referrer-Policy", "no-referrer")
	ctx.Gin.Header("Cache-Control", "no-store")
	app.Response(ctx.Gin, e.SuccessData(result))
}

// GetOwnedBootstrap 加载创建者自己的私人瞬间。
func GetOwnedBootstrap(ctx *contextx.AppContext) {
	e.PanicIfParameterError(ctx.User == nil, "请先登录")
	result, err := plugin_share_service.GetOwnedBootstrap(ctx.Gin.Param("momentId"), *ctx.User)
	bizerr.PanicHTTP(err)
	app.Response(ctx.Gin, e.SuccessData(result))
}

// DeleteShare 物理删除当前用户创建的分享。
func DeleteShare(ctx *contextx.AppContext) {
	e.PanicIfParameterError(ctx.User == nil, "请先登录")
	err := plugin_share_service.DeleteShare(ctx.Gin.Param("shareToken"), *ctx.User)
	bizerr.PanicHTTP(err)
	app.Response(ctx.Gin, e.Success)
}

// ListMyShares 返回当前登录用户创建的分享管理列表。
func ListMyShares(ctx *contextx.AppContext) {
	e.PanicIfParameterError(ctx.User == nil, "请先登录")
	result, err := plugin_share_service.ListMyShares(*ctx.User, plugin_share_service.ShareListQuery{
		Page:     parseQueryInt(ctx.Gin.Query("page"), 1),
		PageSize: parseQueryInt(ctx.Gin.Query("pageSize"), 20),
		Status:   ctx.Gin.Query("status"),
	})
	bizerr.PanicHTTP(err)
	app.Response(ctx.Gin, e.SuccessData(result))
}

// DeleteManagedShare 按分享管理 ID 物理删除当前用户创建的分享。
func DeleteManagedShare(ctx *contextx.AppContext) {
	e.PanicIfParameterError(ctx.User == nil, "请先登录")
	e.PanicIfParameterError(strings.TrimSpace(ctx.Gin.Param("shareId")) == "", "分享 ID 不能为空")
	err := plugin_share_service.DeleteShareByID(ctx.Gin.Param("shareId"), *ctx.User)
	bizerr.PanicHTTP(err)
	app.Response(ctx.Gin, e.Success)
}

// GetResource 输出经过分享作用域校验的插件资源。
func GetResource(ctx *contextx.AppContext) {
	target, err := plugin_share_service.ResolveResource(
		ctx.Gin.Param("shareToken"),
		ctx.Gin.Param("resourceAlias"),
	)
	bizerr.PanicHTTP(err)
	ctx.Gin.Header("Cache-Control", "private, max-age=300")
	ctx.Gin.Header("Referrer-Policy", "no-referrer")
	if target.Mime != "" {
		ctx.Gin.Header("Content-Type", target.Mime)
	}
	ctx.Gin.File(target.Path)
}

func parseQueryInt(raw string, fallback int) int {
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

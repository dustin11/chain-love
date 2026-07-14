package plugin_comment_api

import (
	"errors"
	"senspace/pkg/app"
	"senspace/pkg/app/contextx"
	"senspace/pkg/bizerr"
	"senspace/pkg/e"
	"senspace/service/plugin_comment_service"
)

// ListComments 分页查询插件评论。
func ListComments(ctx *contextx.AppContext) {
	var req plugin_comment_service.ListRequest
	e.PanicParameterError(ctx.Gin.ShouldBindJSON(&req))
	req.ShareToken = ctx.Gin.GetHeader("X-Plugin-Share-Token")
	result, err := plugin_comment_service.ListComments(req, ctx.User)
	panicIfPluginCommentError(err)
	app.Response(ctx.Gin, e.SuccessData(result))
}

// CreateComment 创建插件评论。
func CreateComment(ctx *contextx.AppContext) {
	var req plugin_comment_service.CreateRequest
	e.PanicParameterError(ctx.Gin.ShouldBindJSON(&req))
	req.ShareToken = ctx.Gin.GetHeader("X-Plugin-Share-Token")
	result, err := plugin_comment_service.CreateComment(req, ctx.User)
	panicIfPluginCommentError(err)
	app.Response(ctx.Gin, e.SuccessData(result))
}

// LikeComment 连续点赞插件评论。
func LikeComment(ctx *contextx.AppContext) {
	var req plugin_comment_service.LikeRequest
	e.PanicParameterError(ctx.Gin.ShouldBindJSON(&req))
	req.ShareToken = ctx.Gin.GetHeader("X-Plugin-Share-Token")
	result, err := plugin_comment_service.LikeComment(req, ctx.User)
	panicIfPluginCommentError(err)
	app.Response(ctx.Gin, e.SuccessData(result))
}

// CleanupComments 级联清理插件评论及其点赞。
func CleanupComments(ctx *contextx.AppContext) {
	var req plugin_comment_service.CleanupRequest
	e.PanicParameterError(ctx.Gin.ShouldBindJSON(&req))
	req.ShareToken = ctx.Gin.GetHeader("X-Plugin-Share-Token")
	result, err := plugin_comment_service.CleanupComments(req, ctx.User)
	panicIfPluginCommentError(err)
	app.Response(ctx.Gin, e.SuccessData(result))
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// 透传评论业务错误，避免吞掉点赞上限等业务码。
func panicIfPluginCommentError(err error) {
	if err == nil {
		return
	}
	for _, kind := range []bizerr.Kind{
		bizerr.KindParameter,
		bizerr.KindUnauthorized,
		bizerr.KindForbidden,
		bizerr.KindNotFound,
		bizerr.KindConflict,
	} {
		if bizerr.IsKind(err, kind) {
			bizerr.PanicHTTP(err)
		}
	}

	var serviceErr *e.Error
	if errors.As(err, &serviceErr) {
		panic(serviceErr)
	}

	e.PanicServerErrTipMsg(err, errString(err))
}

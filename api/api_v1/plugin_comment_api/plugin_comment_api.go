package plugin_comment_api

import (
	"senspace/pkg/app"
	"senspace/pkg/app/contextx"
	"senspace/pkg/e"
	"senspace/service/plugin_comment_service"
)

// ListComments 分页查询插件评论。
func ListComments(ctx *contextx.AppContext) {
	var req plugin_comment_service.ListRequest
	e.PanicParameterError(ctx.Gin.ShouldBindJSON(&req))
	result, err := plugin_comment_service.ListComments(req, ctx.User)
	e.PanicServerErrTipMsg(err, errString(err))
	app.Response(ctx.Gin, e.SuccessData(result))
}

// CreateComment 创建插件评论。
func CreateComment(ctx *contextx.AppContext) {
	var req plugin_comment_service.CreateRequest
	e.PanicParameterError(ctx.Gin.ShouldBindJSON(&req))
	result, err := plugin_comment_service.CreateComment(req, ctx.User)
	e.PanicServerErrTipMsg(err, errString(err))
	app.Response(ctx.Gin, e.SuccessData(result))
}

// LikeComment 连续点赞插件评论。
func LikeComment(ctx *contextx.AppContext) {
	var req plugin_comment_service.LikeRequest
	e.PanicParameterError(ctx.Gin.ShouldBindJSON(&req))
	result, err := plugin_comment_service.LikeComment(req, ctx.User)
	e.PanicServerErrTipMsg(err, errString(err))
	app.Response(ctx.Gin, e.SuccessData(result))
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

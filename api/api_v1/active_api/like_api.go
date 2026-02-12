package active_api

import (
	"chain-love/domain/active"
	"chain-love/pkg/app"
	"chain-love/pkg/app/contextx"
	"chain-love/pkg/e"

	"github.com/gin-gonic/gin"
)

// @Summary	点赞
// @Tags		互动
// @Param		data	body		active.Like	true	"id(数据ID), bizType"
// @Security	ApiKeyAuth
// @Router		/api/v1/active/like/add [post]
func LikeAdd(ctx *contextx.AppContext) {
	var form active.Like
	e.PanicIfErr(ctx.Gin.ShouldBindJSON(&form))
	e.PanicIf(form.Id == 0 || form.BizType == 0, "参数错误")

	e.PanicIfErr(form.Add(ctx.User))
	app.Response(ctx.Gin, e.Success)
}

// @Summary	取消/删除点赞
// @Tags		互动
// @Param		data	body		active.Like	true	"id(数据ID), bizType"
// @Security	ApiKeyAuth
// @Router		/api/v1/active/like/del [post]
func LikeDel(ctx *contextx.AppContext) {
	var form active.Like
	e.PanicIfErr(ctx.Gin.ShouldBindJSON(&form))
	e.PanicIf(form.Id == 0 || form.BizType == 0, "参数错误")

	e.PanicIfErr(form.Delete(ctx.User))
	app.Response(ctx.Gin, e.Success)
}

// @Summary	批量获取点赞数
// @Tags		互动
// @Param		data	body		[]active.Like	true	"数组：[{id:1, bizType:1}...]"
// @Router		/api/v1/active/like/info [post]
func LikeGetSummary(ctx *gin.Context) {
	var param []active.Like
	e.PanicIfErr(ctx.ShouldBindJSON(&param))

	list, err := active.GetBatchCounts(param)
	e.PanicIfErr(err)
	app.Response(ctx, e.SuccessData(list))
}

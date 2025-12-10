package ds_api

import (
	"chain-love/domain/ds"
	"chain-love/domain/ds/page"
	"chain-love/pkg/app"
	context "chain-love/pkg/app/contextx"
	"chain-love/pkg/e"
	"chain-love/service/ds_service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Summary	单条数据
// @Tags		房源
// @version	1.0
// @Accept		application/x-json-stream
// @Param		id				path		int		true	"业务id"
// @Param		Authorization	query		string	false	"测试token"
// @Success	200				object		e.Error
// @Failure	500				{object}	e.Error
// @Router		/api/v1/house/info/{id} [get]
func GetHouseInfo(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	e.PanicIfErr(err)
	//id为0时查找登录人草稿状态的数据
	m := ds.House{Id: id}.QueryById()
	app.Response(ctx, e.SuccessData(m))
}

// @Summary	分页列表
// @Tags		房源
// @version	1.0
// @Param		page	query		page.HousePage	true	"房源page对象"
// @Success	200		object		e.Error
// @Failure	500		{object}	e.Error
// @Security	ApiKeyAuth
// @Router		/api/v1/house/page [get]
func GetHousePage(ctx *gin.Context) {
	var page page.HousePage
	var m ds.House
	err := ctx.ShouldBind(&page)
	e.PanicIfErrTipMsg(err, "输入的数据不合法")
	m.GetPage(&page)
	app.Response(ctx, e.SuccessData(page))
}

// @Summary	保存
// @Tags		房源
// @version	1.0
// @Param		name	body		string	true	"如：{'name':'jii','id':8}"
// @Success	200		object		e.Error
// @Failure	500		{object}	e.Error
// @Security	ApiKeyAuth
// @Router		/api/v1/house/save [post]
func SaveHouse(ctx *context.AppContext) {
	var m ds.House
	err := ctx.Gin.ShouldBind(&m)
	e.PanicIfErrTipMsg(err, "输入的数据不合法")

	imgs := ds_service.UploadHouseImg(ctx)
	m.Imgs = imgs
	//m.Save()
	ctx.Response(e.SuccessData(m))
}

// @Summary	删除
// @Tags		房源
// @version	1.0
// @Accept		application/x-json-stream
// @Param		id				path		int		true	"id"
// @Param		Authorization	query		string	false	"测试token"
// @Success	200				object		e.Error
// @Failure	500				{object}	e.Error
// @Router		/api/v1/house/del/{id} [get]
func DeleteHouse(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	e.PanicIfErr(err)
	ds.House{Id: id}.Delete()
	app.Response(ctx, e.Success)
}

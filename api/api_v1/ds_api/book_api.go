package ds_api

import (
	"chain-love/domain/ds"
	"chain-love/domain/ds/page"
	"chain-love/pkg/app"
	"chain-love/pkg/app/contextx"
	"chain-love/pkg/e"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Summary	分页列表
// @Tags		书籍
// @Param		page	query		page.BookPage	true	"查询条件"
// @Success	200		object		e.Error
// @Router		/api/v1/book/page [get]
func BookGetPage(ctx *gin.Context) {
	var p page.BookPage
	var entity ds.Book
	err := ctx.ShouldBind(&p)
	e.PanicIfErrTipMsg(err, "参数错误")
	entity.GetPage(&p)
	app.Response(ctx, e.SuccessData(p))
}

// @Summary	单条数据
// @Tags		书籍
// @Param		id		path		int		true	"ID"
// @Success	200		object		e.Error
// @Router		/api/v1/book/info/{id} [get]
func BookGetById(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	e.PanicIfErr(err)
	m := ds.Book{Id: id}.GetById()
	app.Response(ctx, e.SuccessData(m))
}

// @Summary	保存
// @Tags		书籍
// @Param		data	body		ds.Book	true	"书籍信息"
// @Success	200		object		e.Error
// @Security	ApiKeyAuth
// @Router		/api/v1/book/save [post]
func BookSave(ctx *contextx.AppContext) {
	var model ds.Book
	err := ctx.Gin.ShouldBind(&model)
	e.PanicIfErrTipMsg(err, "参数错误")

	if model.Id > 0 {
		err = model.Update()
	} else {
		err = model.Init(ctx.User).Add()
	}
	e.PanicIfErr(err)
	ctx.Response(e.SuccessData(model))
}

// @Summary	删除
// @Tags		书籍
// @Param		id		path		int		true	"ID"
// @Success	200		object		e.Error
// @Security	ApiKeyAuth
// @Router		/api/v1/book/del/{id} [post]
func BookDel(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	e.PanicIfErr(err)
	ds.Book{Id: id}.Delete()
	app.Response(ctx, e.Success)
}

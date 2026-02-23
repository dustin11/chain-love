package ds_api

import (
	"chain-love/domain/ds"
	"chain-love/domain/ds/page"
	"chain-love/pkg/app"
	"chain-love/pkg/app/contextx"
	"chain-love/pkg/e"
	"chain-love/service/ds_service"
	"chain-love/service/models"
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

// @Summary	根据星球ID返回书ID列表
// @Tags		书籍
// @Param		planetId	path	int	true	"星球ID"
// @Success	200	object	e.Error
// @Router		/api/v1/book/ids/planet/{planetId} [get]
func BookGetIdByPlanetId(ctx *gin.Context) {
	pid, err := strconv.Atoi(ctx.Param("planetId"))
	e.PanicIfErr(err)
	ids, err := ds.Book{}.GetIdByPlanetId(pid)
	if err != nil {
		// 返回空列表或错误
		e.PanicIfErr(err)
	}
	app.Response(ctx, e.SuccessData(map[string][]int{"ids": ids}))
}

// @Summary	保存
// @Tags		书籍
// @Param		data	body		object	true	"书籍信息与分页"
// @Success	200		object		e.Error
// @Security	ApiKeyAuth
// @Router		/api/v1/book/save [post]
func BookSave(ctx *contextx.AppContext) {
	var payload struct {
		Book  ds.Book           `json:"book"`
		Pages []models.PageData `json:"pages"`
	}

	if err := ctx.Gin.ShouldBindJSON(&payload); err != nil {
		e.PanicIfErrTipMsg(err, "参数错误")
	}

	model := &payload.Book

	var err error
	if model.Id > 0 {
		// 获取旧版本号并加1
		oldBook := ds.Book{Id: model.Id}.GetById()
		model.Version = oldBook.Version + 1
		// 更新星球信息，以 JWT 用户为准
		model.PlanetId = ctx.User.PlanetId
		// 传入当前用户 ID 进行校验
		err = model.Update(ctx.User.Id)
	} else {
		// ensure init with user then add
		model = model.Init(ctx.User)
		err = model.Add()
	}
	e.PanicIfErr(err)

	// 保存到文件系统
	err = ds_service.SaveBookFile(model, payload.Pages)
	e.PanicIfServerErrLogMsg(err, "保存书籍文件失败")

	// 仅返回新建/更新后的 id
	app.Response(ctx.Gin, e.SuccessData(map[string]int{"id": model.Id}))
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

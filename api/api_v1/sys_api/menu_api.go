package sys_api

// import (
// 	"chain-love/domain/sys"
// 	page2 "chain-love/domain/sys/page"
// 	"chain-love/pkg/app"
// 	context "chain-love/pkg/app/contextx"
// 	"chain-love/pkg/e"
// 	"chain-love/service/sys_service"
// 	"strconv"

// 	"github.com/gin-gonic/gin"
// )

//	@Summary	单条数据
//	@Tags		菜单
//	@version	1.0
//	@Accept		application/x-json-stream
//	@Param		id				path		int		true	"用户id"
//	@Param		Authorization	query		string	false	"测试token"
//	@Success	200				object		e.Error
//	@Failure	500				{object}	e.Error
//	@Router		/api/v1/menu/one/{id} [get]
// func GetMenuInfo(ctx *gin.Context) {
// 	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
// 	e.PanicIfErr(err)
// 	item := (&sys.Menu{Id: id}).QueryById()
// 	app.Response(ctx, e.SuccessData(item))
// 	//ctx.JSON(http.StatusOK, app.OkResultData(user))
// }

//	@Summary	分页列表
//	@Tags		菜单
//	@version	1.0
//	@Param		page	query		page.MenuPage	true	"菜单"
//	@Success	200		object		e.Error
//	@Failure	500		{object}	e.Error
//	@Security	ApiKeyAuth
//	@Router		/api/v1/menu/page [get]
// func GetPage(ctx *context.AppContext) {
// 	var page page2.MenuPage
// 	var menu sys.Menu
// 	err := ctx.Gin.ShouldBindQuery(&page)
// 	e.PanicIfErrThenMsg(err, "输入的数据不合法")
// 	page.UserId = ctx.User.Id
// 	menu.GetPage(&page)
// 	ctx.Response(e.SuccessData(page))
// 	//app.Response(ctx.Gin, e.SuccessData(page))
// }

//	@Summary	获取菜单及用户权限
//	@Tags		菜单
//	@version	1.0
//	@Accept		application/x-json-stream
//	@Param		Authorization	query		string	false	"测试token"
//	@Success	200				object		e.Error
//	@Failure	500				{object}	e.Error
//	@Router		/api/v1/menu/listPerms [get]
// func GetListWithPerms(ctx *context.AppContext) {
// 	data := sys_service.GetMenuWithPerms(ctx.User.Id)
// 	ctx.Response(e.SuccessData(data))
// }

//	@Summary	保存
//	@Tags		菜单
//	@version	1.0
//	@Param		name	body		string	true	"如：{'name':'jii','id':8}"
//	@Success	200		object		e.Error
//	@Failure	500		{object}	e.Error
//	@Security	ApiKeyAuth
//	@Router		/api/v1/menu/save [post]
// func SaveMenu(ctx *gin.Context) {
// 	var model sys.Menu
// 	err := ctx.ShouldBind(&model)
// 	e.PanicIfErrThenMsg(err, "输入的数据不合法")

// 	sys_service.SaveMenu(model)
// 	app.Response(ctx, e.SuccessData(model))
// }

//	@Summary	删除
//	@Tags		菜单
//	@version	1.0
//	@Accept		application/x-json-stream
//	@Param		id				path		int		true	"用户id"
//	@Param		Authorization	query		string	false	"测试token"
//	@Success	200				object		e.Error
//	@Failure	500				{object}	e.Error
//	@Router		/api/v1/menu/del/{id} [get]
// func DeleteMenu(ctx *gin.Context) {
// 	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
// 	e.PanicIfErr(err)
// 	sys.Menu{Id: id}.Delete()
// 	app.Response(ctx, e.Success)
// }

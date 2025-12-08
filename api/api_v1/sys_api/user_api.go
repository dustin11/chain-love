package sys_api

import (
	"chain-love/domain/sys"
	page2 "chain-love/domain/sys/page"
	"chain-love/pkg/app"
	context "chain-love/pkg/app/contextx"
	"chain-love/pkg/e"
	"chain-love/service/sys_service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// @Summary	分页列表
// @Tags		用户
// @version	1.0
// @Param		page	query		page.UserPage	true	"角色"
// @Success	200		object		e.Error
// @Failure	500		{object}	e.Error
// @Security	ApiKeyAuth
// @Router		/api/v1/user/page [get]
func UserGetPage(ctx *gin.Context) {
	var page page2.UserPage
	var entity sys.User
	err := ctx.ShouldBind(&page)
	e.PanicIfErrThenMsg(err, "输入的数据不合法")
	entity.GetPage(&page)
	app.Response(ctx, e.SuccessData(page))
}

// @Summary	单条数据
// @Tags		用户
// @version	1.0
// @Accept		application/x-json-stream
// @Param		id				path		int		true	"用户id"
// @Param		Authorization	query		string	false	"测试token"
// @Success	200				object		e.Error
// @Failure	500				{object}	e.Error
// @Router		/api/v1/user/info/{id} [get]
func UserGetById(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	e.PanicIfErr(err)
	m := sys.User{Id: id}.QueryById()
	//roleIds := sys_service.GetRoleIds(m.Id)
	//m.RoleIds = roleIds
	app.Response(ctx, e.SuccessData(m))
}

// @Summary	单条数据
// @Tags		用户
// @version	1.0
// @Accept		application/x-json-stream
// @Param		addr			path		string	true	"用户addr"
// @Param		Authorization	query		string	false	"测试token"
// @Success	200				object		e.Error
// @Failure	500				{object}	e.Error
// @Router		/api/v1/user/addr/{addr} [get]
func UserGetByAddr(ctx *gin.Context) {
	addr := ctx.Param("addr")
	e.PanicIf(addr == "", "addr不能为空！")
	m := sys.User{Addr: addr}.GetByAddr()
	app.Response(ctx, e.SuccessData(m))
}

// @Summary	保存
// @Tags		用户
// @version	1.0
// @Param		name	body		string	true	"如：{'name':'jii','id':8}"
// @Success	200		object		e.Error
// @Failure	500		{object}	e.Error
// @Security	ApiKeyAuth
// @Router		/api/v1/user/save [post]
func UserSave(ctx *context.AppContext) {
	var model sys.User
	err := ctx.Gin.ShouldBind(&model)
	e.PanicIfErrThenMsg(err, "输入的数据不合法")

	sys_service.UserSave(&model)
	ctx.Response(e.SuccessData(model))
}

// @Summary	删除
// @Tags		货品
// @version	1.0
// @Accept		application/x-json-stream
// @Param		id				path		int		true	"用户id"
// @Param		Authorization	query		string	false	"测试token"
// @Success	200				object		e.Error
// @Failure	500				{object}	e.Error
// @Router		/api/v1/user/del/{id} [get]
func UserDel(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	e.PanicIfErr(err)
	//删除角色
	sys.UserRole{UserId: id}.DeleteByUserId()
	//删除用户
	sys.User{Id: id}.Delete()
	app.Response(ctx, e.Success)
}

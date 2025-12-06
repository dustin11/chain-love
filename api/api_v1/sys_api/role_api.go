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

// @Summary 单条数据
// @Tags 角色
// @version 1.0
// @Accept application/x-json-stream
// @Param id path int true "用户id"
// @Param Authorization query string false "测试token"
// @Success 200 object e.Error
// @Failure 500 {object} e.Error
// @Router /api/v1/role/one/{id} [get]
func GetRoleInfo(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	e.PanicIfErr(err)
	m := sys.Role{Id: id}.QueryById()
	menuIds := sys_service.GetMenuIds(id)
	m.MenuIds = menuIds
	app.Response(ctx, e.SuccessData(m))
}

// @Summary 分页列表
// @Tags 角色
// @version 1.0
// @Param page query page.RolePage true "角色"
// @Success 200 object e.Error
// @Failure 500 {object} e.Error
// @Security ApiKeyAuth
// @Router /api/v1/role/page [get]
func GetRolePage(ctx *gin.Context) {
	var page page2.RolePage
	var role sys.Role
	err := ctx.ShouldBind(&page)
	e.PanicIfErrThenMsg(err, "输入的数据不合法")
	role.GetPage(&page)
	app.Response(ctx, e.SuccessData(page))
}

// @Summary 保存
// @Tags 角色
// @version 1.0
// @Param name body string true "如：{'name':'jii','id':8}"
// @Success 200 object e.Error
// @Failure 500 {object} e.Error
// @Security ApiKeyAuth
// @Router /api/v1/role/save [post]
func SaveRole(ctx *context.AppContext) {
	var model sys.Role
	err := ctx.Gin.ShouldBind(&model)
	e.PanicIfErrThenMsg(err, "输入的数据不合法")
	sys_service.SaveRole(model)
	ctx.Response(e.SuccessData(model))
}

// @Summary 删除
// @Tags 角色
// @version 1.0
// @Accept application/x-json-stream
// @Param id path int true "id"
// @Param Authorization query string false "测试token"
// @Success 200 object e.Error
// @Failure 500 {object} e.Error
// @Router /api/v1/role/del/{id} [get]
func DeleteRole(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	e.PanicIfErr(err)
	userRole := sys.UserRole{RoleId: id}.GetByRoleId()
	e.PanicIf(len(userRole) > 0, "角色使用中，无法删除！")
	sys.Role{Id: id}.Delete()
	app.Response(ctx, e.Success)
}

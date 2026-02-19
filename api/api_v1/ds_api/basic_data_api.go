package ds_api

import (
	"chain-love/domain/ds/enum"
	"chain-love/pkg/app"
	"chain-love/pkg/e"

	"github.com/gin-gonic/gin"
)

// @Summary	装修类型
// @Tags		基础数据
// @Accept		application/x-json-stream
// @Param		Authorization	query		string	false	"测试token"
// @Success	200				object		e.Error
// @Failure	500				{object}	e.Error
// @Router		/api/v1/basic/decorationType [get]
func GetDecorationType(ctx *gin.Context) {
	var m = enum.Decoration_Rough
	list := enum.GetList(m)
	app.Response(ctx, e.SuccessData(list))
}

// @Summary	房屋设施
// @Tags		基础数据
// @Accept		application/x-json-stream
// @Param		Authorization	query		string	false	"测试token"
// @Success	200				object		e.Error
// @Failure	500				{object}	e.Error
// @Router		/api/v1/basic/furnitureType [get]
func GetFurnitureType(ctx *gin.Context) {
	var m = enum.Furniture_Heater
	list := enum.GetList(m)
	app.Response(ctx, e.SuccessData(list))
}

// @Summary	房源类型
// @Tags		基础数据
// @Param		Authorization	query		string	false	"测试token"
// @Success	200				object		e.Error
// @Failure	500				{object}	e.Error
// @Router		/api/v1/basic/houseType [get]
func GetHouseType(ctx *gin.Context) {
	var m = enum.House_Type_House
	list := enum.GetList(m)
	app.Response(ctx, e.SuccessData(list))
}

// @Summary	朝向
// @Tags		基础数据
// @Param		Authorization	query		string	false	"测试token"
// @Success	200				object		e.Error
// @Failure	500				{object}	e.Error
// @Router		/api/v1/basic/orientationType [get]
func GetOrientationType(ctx *gin.Context) {
	var m = enum.Orientation_East
	list := enum.GetList(m)
	app.Response(ctx, e.SuccessData(list))
}

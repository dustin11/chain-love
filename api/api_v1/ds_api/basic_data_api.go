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

// @Summary	租金包含
// @Tags		基础数据
// @Param		Authorization	query		string	false	"测试token"
// @Success	200				object		e.Error
// @Failure	500				{object}	e.Error
// @Router		/api/v1/basic/rentInclude [get]
func GetRentInclude(ctx *gin.Context) {
	var m = enum.Rent_Include_CableTV
	list := enum.GetList(m)
	app.Response(ctx, e.SuccessData(list))
}

// @Summary	付款方式
// @Tags		基础数据
// @Param		Authorization	query		string	false	"测试token"
// @Success	200				object		e.Error
// @Failure	500				{object}	e.Error
// @Router		/api/v1/basic/rentPayType [get]
func GetRentPayType(ctx *gin.Context) {
	var m = enum.Rent_Pay_1To1
	list := enum.GetList(m)
	app.Response(ctx, e.SuccessData(list))
}

// @Summary	出租状态
// @Tags		基础数据
// @Param		Authorization	query		string	false	"测试token"
// @Success	200				object		e.Error
// @Failure	500				{object}	e.Error
// @Router		/api/v1/basic/rentState [get]
func GetRentState(ctx *gin.Context) {
	var m = enum.Rent_State_NoPublish
	list := enum.GetList(m)
	app.Response(ctx, e.SuccessData(list))
}

// @Summary	出租类型
// @Tags		基础数据
// @Param		Authorization	query		string	false	"测试token"
// @Success	200				object		e.Error
// @Failure	500				{object}	e.Error
// @Router		/api/v1/basic/rentType [get]
func GetRentType(ctx *gin.Context) {
	var m = enum.Rent_Type_House
	list := enum.GetList(m)
	app.Response(ctx, e.SuccessData(list))
}

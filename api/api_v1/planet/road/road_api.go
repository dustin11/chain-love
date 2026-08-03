package road_api

import (
	"strconv"

	"senspace/pkg/app"
	"senspace/pkg/app/contextx"
	"senspace/pkg/bizerr"
	"senspace/pkg/e"
	road_service "senspace/service/planet/road"

	"github.com/gin-gonic/gin"
)

// GetPublished 公开读取 planet 道路网络。
func GetPublished(c *gin.Context) {
	planetId, err := strconv.Atoi(c.Param("planetId"))
	e.PanicIfParameterError(err != nil || planetId <= 0, "planetId无效")
	document, err := road_service.GetPublished(planetId)
	if err != nil {
		bizerr.PanicHTTP(err)
	}
	app.Response(c, e.SuccessData(document))
}

// SavePublished 由星球主人发布 planet 道路网络。
func SavePublished(c *contextx.AppContext) {
	planetId, err := strconv.Atoi(c.Gin.Param("planetId"))
	e.PanicIfParameterError(err != nil || planetId <= 0, "planetId无效")
	var request road_service.SaveRequest
	e.PanicParameterError(c.Gin.ShouldBindJSON(&request))
	document, err := road_service.SavePublished(planetId, request, c.User)
	if err != nil {
		bizerr.PanicHTTP(err)
	}
	app.Response(c.Gin, e.SuccessData(document))
}

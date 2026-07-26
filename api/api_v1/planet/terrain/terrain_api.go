package terrain_api

import (
	"strconv"

	"senspace/pkg/app"
	"senspace/pkg/app/contextx"
	"senspace/pkg/bizerr"
	"senspace/pkg/e"
	terrain_service "senspace/service/planet/terrain"

	"github.com/gin-gonic/gin"
)

// GetPublished 公开读取 planet 地形。
func GetPublished(c *gin.Context) {
	planetId, err := strconv.Atoi(c.Param("planetId"))
	e.PanicIfParameterError(err != nil || planetId <= 0, "planetId无效")
	document, err := terrain_service.GetPublished(planetId)
	if err != nil {
		bizerr.PanicHTTP(err)
	}
	app.Response(c, e.SuccessData(document))
}

// SavePublished 由星球主人发布 planet 地形。
func SavePublished(c *contextx.AppContext) {
	planetId, err := strconv.Atoi(c.Gin.Param("planetId"))
	e.PanicIfParameterError(err != nil || planetId <= 0, "planetId无效")
	var request terrain_service.SaveRequest
	e.PanicParameterError(c.Gin.ShouldBindJSON(&request))
	document, err := terrain_service.SavePublished(planetId, request, c.User)
	if err != nil {
		bizerr.PanicHTTP(err)
	}
	app.Response(c.Gin, e.SuccessData(document))
}

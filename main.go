package main

import (
	"chain-love/asset"
	_ "chain-love/docs"
	"chain-love/domain"
	"chain-love/domain/d_util"
	"chain-love/pkg/logging"
	"chain-love/pkg/setting"
	"chain-love/pkg/setting/consts"
	"chain-love/pkg/util"
	"chain-love/routers"
	"log"

	"github.com/gin-gonic/gin"
)

func init() {
	if consts.Getenv() != "dev" {
		log.Println("=============>env:" + consts.Getenv())
		gin.SetMode(gin.ReleaseMode)
		asset.Restore()
	}
	setting.Setup()
	logging.Setup()
	// ensure database exists before opening connection (requires CREATE DATABASE privilege)
	if err := d_util.EnsureDatabaseExists(setting.Config.Database.Name); err != nil {
		logging.Error("ensure database exists error:", err)
		return
	}
	domain.Setup()
	d_util.InitTable(domain.Db)
	util.Setup()
}

//	@title						Gin swagger
//	@version					1.0
//	@description				Gin swagger 示例项目
//	@securityDefinitions.apikey	ApiKeyAuth
//	@in							headerl
//	@name						Authorization
//	@BasePath					/

// @license.name	Apache 2.0
// @license.url	http://www.apache.org/licenses/LICENSE-2.0.html
func main() {
	router := routers.SetupRouter()
	_ = router.Run(":" + setting.Config.Server.Port)
}

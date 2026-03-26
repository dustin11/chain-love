package main

import (
	"senspace/asset"
	_ "senspace/docs"
	"senspace/domain"
	"senspace/domain/d_util"
	"senspace/pkg/i18n"
	"senspace/pkg/logging"
	"senspace/pkg/setting"
	"senspace/pkg/setting/consts"
	"senspace/pkg/util"
	"senspace/routers"
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
	i18n.Setup() // 初始化多语言
	// ensure database exists before opening connection (requires CREATE DATABASE privilege)
	if err := d_util.EnsureDatabaseExists(setting.Config.Database.Name); err != nil {
		logging.Error("ensure database exists error:", err)
		return
	}
	domain.Setup()
	d_util.InitTable(domain.Db)
	util.Setup()
}

//	@title						Senspace API
//	@version					1.0
//	@description				Senspace API service
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

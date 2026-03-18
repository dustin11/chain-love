package dev_api

import (
	"chain-love/domain/dev"
	"chain-love/pkg/app"
	"chain-love/pkg/app/contextx"
	"chain-love/pkg/e"
	"chain-love/service/dev_service"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetPluginList(c *gin.Context) {
	var plugin dev.Plugin
	list := plugin.List()
	for i := range list {
		idStr := strconv.FormatInt(list[i].Id, 10)
		v := dev_service.GetLatestVersion(idStr)
		if v != "" {
			list[i].Version = v
			// list[i].Name = list[i].Name + " v" + v
		}
	}
	app.Response(c, e.SuccessData(list))
}

func GetPluginTree(c *gin.Context) {
	pluginId := c.Param("pluginId")
	tree, err := dev_service.GetPluginTree(pluginId)
	e.PanicServerErr(err)
	app.Response(c, e.SuccessData(tree))
}

func UploadFile(c *gin.Context) {
	pluginId := c.Query("pluginId")
	path := c.Query("path")
	file, err := c.FormFile("file")
	e.PanicParameterErrorTipMsg(err, "file missing")
	err = dev_service.UploadFile(pluginId, path, file)
	e.PanicServerErr(err)
	app.Response(c, e.Success)
}

func AddFolder(c *gin.Context) {
	var req struct {
		PluginId string `json:"pluginId"`
		Path     string `json:"path"`
	}
	err := c.ShouldBindJSON(&req)
	e.PanicParameterError(err)
	err = dev_service.AddFolder(req.PluginId, req.Path)
	e.PanicServerErr(err)
	app.Response(c, e.Success)
}

func Rename(c *gin.Context) {
	var req struct {
		PluginId string `json:"pluginId"`
		OldPath  string `json:"oldPath"`
		NewName  string `json:"newName"`
	}
	err := c.ShouldBindJSON(&req)
	e.PanicParameterError(err)
	err = dev_service.Rename(req.PluginId, req.OldPath, req.NewName)
	e.PanicServerErrTipMsg(err, "重命名失败")
	app.Response(c, e.Success)
}

func Delete(c *gin.Context) {
	var req struct {
		PluginId string `json:"pluginId"`
		Path     string `json:"path"`
	}
	err := c.ShouldBindJSON(&req)
	e.PanicParameterError(err)
	err = dev_service.Delete(req.PluginId, req.Path)
	e.PanicServerErr(err)
	app.Response(c, e.Success)
}

func DeletePlugin(c *gin.Context) {
	var req struct {
		PluginId string `json:"pluginId"`
	}
	err := c.ShouldBindJSON(&req)
	e.PanicParameterError(err)
	err = dev_service.DeletePlugin(req.PluginId)
	e.PanicServerErr(err)
	app.Response(c, e.Success)
}

func SavePlugin(c *contextx.AppContext) {
	pluginId := c.Gin.Query("pluginId")
	form, err := c.Gin.MultipartForm()
	e.PanicParameterErrorTipMsg(err, "invalid multipart form")

	res, err := dev_service.SavePlugin(c, pluginId, form)
	e.PanicServerErr(err)
	app.Response(c.Gin, e.SuccessData(res))
}

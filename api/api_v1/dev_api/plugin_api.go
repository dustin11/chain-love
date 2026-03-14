package dev_api

import (
	"chain-love/domain/dev"
	"chain-love/pkg/app"
	"chain-love/pkg/app/contextx"
	"chain-love/pkg/e"
	"chain-love/service/dev_service"

	"github.com/gin-gonic/gin"
)

func GetPluginList(c *gin.Context) {
	var plugin dev.Plugin
	list := plugin.List()
	app.Response(c, e.SuccessData(list))
}

func GetPluginTree(c *gin.Context) {
	pluginId := c.Param("pluginId")
	tree, err := dev_service.GetPluginTree(pluginId)
	if err != nil {
		app.Response(c, e.ParameterError(err.Error()))
		return
	}
	app.Response(c, e.SuccessData(tree))
}

func UploadFile(c *gin.Context) {
	pluginId := c.Query("pluginId")
	path := c.Query("path")
	file, err := c.FormFile("file")
	if err != nil {
		app.Response(c, e.ParameterError("file missing"))
		return
	}
	err = dev_service.UploadFile(pluginId, path, file)
	if err != nil {
		app.Response(c, e.ParameterError(err.Error()))
		return
	}
	app.Response(c, e.Success)
}

func AddFolder(c *gin.Context) {
	var req struct {
		PluginId string `json:"pluginId"`
		Path     string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Response(c, e.ParameterError("invalid request"))
		return
	}
	if err := dev_service.AddFolder(req.PluginId, req.Path); err != nil {
		app.Response(c, e.ParameterError(err.Error()))
		return
	}
	app.Response(c, e.Success)
}

func Rename(c *gin.Context) {
	var req struct {
		PluginId string `json:"pluginId"`
		OldPath  string `json:"oldPath"`
		NewName  string `json:"newName"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Response(c, e.ParameterError("invalid request"))
		return
	}
	if err := dev_service.Rename(req.PluginId, req.OldPath, req.NewName); err != nil {
		app.Response(c, e.ParameterError(err.Error()))
		return
	}
	app.Response(c, e.Success)
}

func Delete(c *gin.Context) {
	var req struct {
		PluginId string `json:"pluginId"`
		Path     string `json:"path"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		app.Response(c, e.ParameterError("invalid request"))
		return
	}
	if err := dev_service.Delete(req.PluginId, req.Path); err != nil {
		app.Response(c, e.ParameterError(err.Error()))
		return
	}
	app.Response(c, e.Success)
}

func SavePlugin(c *contextx.AppContext) {
	pluginId := c.Gin.Query("pluginId")
	form, err := c.Gin.MultipartForm()
	if err != nil {
		app.Response(c.Gin, e.ParameterError("invalid multipart form"))
		return
	}

	res, err := dev_service.SavePlugin(c, pluginId, form)
	if err != nil {
		app.Response(c.Gin, e.ParameterError(err.Error()))
		return
	}
	app.Response(c.Gin, e.SuccessData(res))
}

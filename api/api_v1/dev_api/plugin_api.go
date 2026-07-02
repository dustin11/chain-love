package dev_api

import (
	"senspace/domain/dev"
	"senspace/pkg/app"
	"senspace/pkg/app/contextx"
	"senspace/pkg/bizerr"
	"senspace/pkg/e"
	"senspace/pkg/setting"
	"senspace/pkg/util"
	"senspace/service/dev_service"
	"strconv"

	"github.com/gin-gonic/gin"
)

const defaultPluginUploadMaxSize = 30 * 1024 * 1024

func resolvePluginUploadMaxSize() int64 {
	if setting.Config.App.PluginUploadMaxSize > 0 {
		return setting.Config.App.PluginUploadMaxSize
	}
	return defaultPluginUploadMaxSize
}

func preparePluginMultipartBody(c *gin.Context) {
	err := util.LimitRequestBodyBytes(c.Writer, c.Request, resolvePluginUploadMaxSize())
	if util.IsRequestBodyTooLargeError(err) {
		e.PanicParameterErrorTipMsg(err, "上传内容超过 30 MB 限制")
	}
	e.PanicServerErr(err)
}

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

func GetPluginVersions(c *gin.Context) {
	pluginId := c.Param("pluginId")
	versions, err := dev_service.ListPluginVersions(pluginId)
	e.PanicServerErr(err)
	app.Response(c, e.SuccessData(versions))
}

func GetPluginTree(c *gin.Context) {
	pluginId := c.Param("pluginId")
	version := c.Query("version")
	tree, err := dev_service.GetPluginTree(pluginId, version)
	e.PanicServerErr(err)
	app.Response(c, e.SuccessData(tree))
}

func UploadFile(c *gin.Context) {
	pluginId := c.Query("pluginId")
	path := c.Query("path")
	preparePluginMultipartBody(c)
	file, err := c.FormFile("file")
	if util.IsRequestBodyTooLargeError(err) {
		e.PanicParameterErrorTipMsg(err, "上传内容超过 30 MB 限制")
	}
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

func DeletePlugin(c *contextx.AppContext) {
	var req struct {
		PluginId string `json:"pluginId"`
	}
	e.PanicIfUnauthorizedErr(c.User == nil, "无授权信息！")
	err := c.Gin.ShouldBindJSON(&req)
	e.PanicParameterError(err)
	err = dev_service.DeletePlugin(c.User, req.PluginId)
	bizerr.PanicHTTP(err)
	app.Response(c.Gin, e.Success)
}

func SavePlugin(c *contextx.AppContext) {
	pluginId := c.Gin.Query("pluginId")
	preparePluginMultipartBody(c.Gin)
	form, err := c.Gin.MultipartForm()
	if util.IsRequestBodyTooLargeError(err) {
		e.PanicParameterErrorTipMsg(err, "上传内容超过 30 MB 限制")
	}
	e.PanicParameterErrorTipMsg(err, "invalid multipart form")

	res, err := dev_service.SavePlugin(c, pluginId, form)
	e.PanicServerErr(err)
	app.Response(c.Gin, e.SuccessData(res))
}

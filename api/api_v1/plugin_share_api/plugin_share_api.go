package plugin_share_api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"senspace/pkg/app"
	"senspace/pkg/app/contextx"
	"senspace/pkg/bizerr"
	"senspace/pkg/e"
	"senspace/service/plugin_share_service"
)

const maxShareBackgroundBytes = 12 * 1024 * 1024

// CreateShare 创建单插件分享快照。
func CreateShare(ctx *contextx.AppContext) {
	e.PanicIfParameterError(ctx.User == nil, "请先登录")
	payload := ctx.Gin.PostForm("payload")
	e.PanicIfParameterError(strings.TrimSpace(payload) == "", "payload不能为空")
	var input plugin_share_service.CreateInput
	e.PanicParameterError(json.Unmarshal([]byte(payload), &input))
	file, err := ctx.Gin.FormFile("background")
	e.PanicParameterErrorTipMsg(err, "分享背景不能为空")
	e.PanicIfParameterError(file.Size <= 0 || file.Size > maxShareBackgroundBytes, "分享背景大小无效")
	opened, err := file.Open()
	e.PanicServerErr(err)
	defer func() { _ = opened.Close() }()
	background, err := io.ReadAll(io.LimitReader(opened, maxShareBackgroundBytes+1))
	e.PanicServerErr(err)
	e.PanicIfParameterError(len(background) > maxShareBackgroundBytes, "分享背景超过大小限制")
	extension, err := detectBackgroundExtension(background)
	e.PanicParameterError(err)
	result, err := plugin_share_service.CreateShare(*ctx.User, input, background, extension)
	bizerr.PanicHTTP(err)
	app.Response(ctx.Gin, e.SuccessData(result))
}

// GetBootstrap 返回分享页最小公开启动数据。
func GetBootstrap(ctx *contextx.AppContext) {
	result, err := plugin_share_service.GetBootstrap(ctx.Gin.Param("shareToken"), ctx.User)
	bizerr.PanicHTTP(err)
	ctx.Gin.Header("Referrer-Policy", "no-referrer")
	ctx.Gin.Header("Cache-Control", "no-store")
	app.Response(ctx.Gin, e.SuccessData(result))
}

// RevokeShare 撤销当前用户创建的分享。
func RevokeShare(ctx *contextx.AppContext) {
	e.PanicIfParameterError(ctx.User == nil, "请先登录")
	err := plugin_share_service.RevokeShare(ctx.Gin.Param("shareToken"), *ctx.User)
	bizerr.PanicHTTP(err)
	app.Response(ctx.Gin, e.Success)
}

// GetResource 输出经过分享作用域校验的插件资源。
func GetResource(ctx *contextx.AppContext) {
	target, err := plugin_share_service.ResolveResource(
		ctx.Gin.Param("shareToken"),
		ctx.Gin.Param("resourceAlias"),
	)
	bizerr.PanicHTTP(err)
	ctx.Gin.Header("Cache-Control", "private, max-age=300")
	ctx.Gin.Header("Referrer-Policy", "no-referrer")
	if target.Mime != "" {
		ctx.Gin.Header("Content-Type", target.Mime)
	}
	ctx.Gin.File(target.Path)
}

func detectBackgroundExtension(data []byte) (string, error) {
	mimeType := http.DetectContentType(data)
	switch mimeType {
	case "image/webp":
		return ".webp", nil
	case "image/png":
		return ".png", nil
	case "image/jpeg":
		return ".jpg", nil
	default:
		return "", &unsupportedBackgroundError{}
	}
}

type unsupportedBackgroundError struct{}

func (e *unsupportedBackgroundError) Error() string {
	return "分享背景必须是 WebP、PNG 或 JPEG 图片"
}

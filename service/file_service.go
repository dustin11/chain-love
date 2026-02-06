package service

import (
	"bytes"
	context "chain-love/pkg/app/contextx"
	"chain-love/pkg/e"
	"chain-love/pkg/setting"
	"chain-love/pkg/util"
	"fmt"
	"image"
	"image/gif"
	"image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

// UploadFormFiles 保存表单中所有文件（返回 map[fieldName]savedFileName）
func UploadFormFiles(ctx *context.AppContext) map[string]string {
	form, err := ctx.Gin.MultipartForm()
	e.PanicIfErr(err)

	// 总文件数限制
	total := 0
	for _, files := range form.File {
		total += len(files)
	}
	e.PanicIf(total == 0, "上传文件为空！")
	e.PanicIf(total > 35, "最多上传35张图片")

	filesMap := make(map[string]string)
	for field, files := range form.File {
		// 只取第一个文件（前端每个 field 只会传一个）
		if len(files) == 0 {
			continue
		}
		f := files[0]
		e.PanicIf(f == nil, "上传文件为空！")
		e.PanicIf(f.Size > setting.Config.App.ImageMaxSize, "图片超过最大限制！")

		// 读取文件字节并关闭
		src, err := f.Open()
		e.PanicIfErr(err)
		data, err := io.ReadAll(src)
		src.Close()
		e.PanicIfErr(err)
		e.PanicIf(len(data) == 0, "上传文件为空！")

		// MIME sniff（前512字节）
		limit := 512
		if len(data) < limit {
			limit = len(data)
		}
		ctype := http.DetectContentType(data[:limit])
		e.PanicIf(!strings.HasPrefix(ctype, "image/"), "不是图片文件")

		// 检查图像配置（格式、尺寸）
		cfg, format, err := image.DecodeConfig(bytes.NewReader(data))
		e.PanicIfErr(err)
		e.PanicIf(cfg.Width > 10000 || cfg.Height > 10000, "图片分辨率太大")

		if exists, _ := util.Contain(format, setting.Config.App.ImageExts); !exists {
			e.PanicIf(true, fmt.Sprintf("只允许上传%s格式的文件!", strings.Join(setting.Config.App.ImageExts, ",")))
		}

		// 确保目录存在
		dir := filepath.Join(setting.Config.App.FilePath.Image, strconv.Itoa(int(ctx.User.Id)))
		util.CreateDirIfNotExits(dir)

		// 重新解码并统一处理（去元数据、缩放等），然后保存
		upImg, err := imaging.Decode(bytes.NewReader(data))
		e.PanicIfErr(err)

		if upImg.Bounds().Size().X > 1200 || upImg.Bounds().Size().Y > 1200 {
			upImg = imaging.Resize(upImg, 1000, 0, imaging.CatmullRom)
		}
		// 若原为 png 且无透明则转为 jpeg（节省空间）
		if format == "png" && !hasAlpha(upImg) {
			format = "jpg"
		}
		fileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), "."+format)
		full := filepath.Join(dir, fileName)
		out, err := os.Create(full)
		e.PanicIfErr(err)
		// 根据决定的格式进行编码并保存
		switch format {
		case "jpg":
			e.PanicIfErr(imaging.Encode(out, upImg, imaging.JPEG, imaging.JPEGQuality(80)))
		case "gif":
			e.PanicIfErr(gif.Encode(out, upImg, nil))
		case "png":
			enc := png.Encoder{CompressionLevel: png.BestCompression}
			e.PanicIfErr(enc.Encode(out, upImg))
		default:
			e.PanicIfErr(imaging.Encode(out, upImg, imaging.JPEG, imaging.JPEGQuality(80)))
		}
		e.PanicIfErr(out.Close())

		// 只返回文件名（不要带 uid 前缀）
		filesMap[field] = fileName
	}

	return filesMap
}

// DeleteImageFile 删除磁盘上的图片文件（uid 是上传用户 id，img 是文件名）
func DeleteImageFile(uid string, img string) {
	imgPath := fmt.Sprintf("%s%s/%s", setting.Config.App.FilePath.Image, uid, img)
	err := util.RemoveIfExists(imgPath)
	e.PanicIfErr(err)
}

// helper: 检测是否包含透明像素（简单采样）
func hasAlpha(img image.Image) bool {
	switch im := img.(type) {
	case *image.NRGBA:
		for i := 3; i < len(im.Pix); i += 4 {
			if im.Pix[i] != 0xFF {
				return true
			}
		}
		return false
	case *image.NRGBA64:
		for i := 3; i < len(im.Pix); i += 4 {
			if im.Pix[i] != 0xFF {
				return true
			}
		}
		return false
	default:
		b := img.Bounds()
		for y := b.Min.Y; y < b.Max.Y; y += max(1, (b.Dy() / 10)) {
			for x := b.Min.X; x < b.Max.X; x += max(1, (b.Dx() / 10)) {
				_, _, _, a := img.At(x, y).RGBA()
				if a != 0xFFFF {
					return true
				}
			}
		}
		return false
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

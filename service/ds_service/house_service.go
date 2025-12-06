package ds_service

import (
	context "chain-love/pkg/app/contextx"
	"chain-love/pkg/e"
	"chain-love/pkg/setting"
	"chain-love/pkg/util"
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

// 添加图片
func UploadHouseImg(ctx *context.AppContext) string {
	//处理文件夹
	filePath := fmt.Sprintf("%s%d/", setting.Config.App.ImageSavePath, ctx.User.Id)
	util.CreateDirIfNotExits(filePath)
	//获取文件
	//ctx.Gin.Request.MultipartForm.File["imgfile"]
	f, _ := ctx.Gin.FormFile("imgfile")
	e.PanicIf(f == nil, "上传文件为空！")
	//检查大小
	e.PanicIf(f.Size > setting.Config.App.ImageMaxSize, "图片超过最大限制！")

	//for _, f := range files {
	fileExt := strings.ToLower(path.Ext(f.Filename))
	if exits, _ := util.Contain(fileExt, setting.Config.App.ImageExts); !exits {
		e.PanicMsg(fmt.Sprintf("只允许上传%s格式的文件!", strings.Join(setting.Config.App.ImageExts, ",")))
	}
	fileName := fmt.Sprintf("%s%s", time.Now().Format("20060102150405"), fileExt)

	//图片处理（实现上传文件直接缩放，不需要先存磁盘 折腾了很长时间）
	file, err := f.Open()
	e.PanicIfErr(err)
	defer file.Close()

	upImg, e1 := imaging.Decode(file)
	e.PanicIfErr(e1)
	if upImg.Bounds().Size().X > 1000 || upImg.Bounds().Size().Y > 1000 {
		upImg = imaging.Resize(upImg, 800, 0, imaging.CatmullRom)
	}
	err = imaging.Save(upImg, filePath+fileName)
	if err != nil {
		log.Fatalf("failed to save image: %v", err)
	}
	//缩略图
	//sImg := imaging.Resize(upImg, 128, 128, imaging.Lanczos)
	//ctx.Gin.SaveUploadedFile(f, filePath+fileName)
	//}

	return fileName
}

func DelHouseImg(img string, uid string) {
	imgPath := fmt.Sprintf("%s%s/%s", setting.Config.App.ImageSavePath, uid, img)
	err := os.Remove(imgPath)
	e.PanicIfErr(err)
}

package ds_api

import (
	"chain-love/domain/ds"
	"chain-love/domain/ds/page"
	"chain-love/pkg/app"
	"chain-love/pkg/app/contextx"
	"chain-love/pkg/e"
	"chain-love/pkg/setting"
	"chain-love/service"
	"encoding/json"
	"path"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// @Summary	上传图片（支持多图，最多35张）
// @Tags		图片
// @Param		bizType	formData	string	false	"业务类型"
// @Param		bizSubType	formData	string	false	"业务子类"
// @Param		imgfile	formData	file	true	"图片（可多次提交多个 imgfile 字段）"
// @Security	ApiKeyAuth
// @Router		/api/v1/image/save [post]
// 支持 items+files 的批量接口：items 为 JSON 数组，file key 对应文件字段名（如 file123）
func ImageSave(ctx *contextx.AppContext) {
	bizType := uint8(0)
	if s := ctx.Gin.PostForm("bizType"); s != "" {
		v, err := strconv.ParseUint(s, 10, 8)
		e.PanicIfErr(err)
		bizType = uint8(v)
	}
	bizSubType := uint8(0)
	if s := ctx.Gin.PostForm("bizSubType"); s != "" {
		v, err := strconv.ParseUint(s, 10, 8)
		e.PanicIfErr(err)
		bizSubType = uint8(v)
	}

	itemsStr := ctx.Gin.PostForm("items")
	e.PanicIf(itemsStr == "", "图片信息不能为空！")

	var items []struct {
		Id  uint64    `json:"id"`
		Pos []float64 `json:"pos"`
		// localId 由前端的 id 字段直接回传
	}
	e.PanicIfErr(json.Unmarshal([]byte(itemsStr), &items))

	uploaded := service.UploadFormFiles(ctx) // map[fieldName]fileName

	// 收集待批量插入的新记录及其 localId，仅返回新建记录
	var newImgs []ds.Image
	var newLocalIds []int
	var created []map[string]interface{}

	for _, it := range items {
		if it.Id < 36 {
			// new item，文件字段名约定 file{localId}
			key := "file" + strconv.Itoa(int(it.Id))
			fileName, ok := uploaded[key]
			e.PanicIf(!ok, "缺失文件: "+key)
			url := path.Join(strconv.Itoa(int(ctx.User.Id)), fileName)
			styleB, _ := json.Marshal(map[string]interface{}{"pos": it.Pos})
			img := ds.Image{
				Url:        url,
				BizType:    bizType,
				BizSubType: bizSubType,
				Style:      string(styleB),
			}
			img.Init(ctx.User)
			newImgs = append(newImgs, img)
			newLocalIds = append(newLocalIds, int(it.Id))
		} else {
			// existing: 只更新 Style 字段
			styleB, _ := json.Marshal(map[string]interface{}{"pos": it.Pos})
			e.PanicIfErr(ds.UpdateStyle(it.Id, ctx.User.Id, string(styleB)))
		}
	}

	// 批量插入新记录（若有），返回新建项并带回 localId
	if len(newImgs) > 0 {
		e.PanicIfErr(ds.AddBatch(newImgs))
		for j, img := range newImgs {
			created = append(created, map[string]interface{}{
				"localId": newLocalIds[j],
				"id":      img.Id,
				"url":     img.Url,
				"style":   img.Style,
			})
		}
	}

	app.Response(ctx.Gin, e.SuccessData(created))
}

// @Summary	分页列表
// @Tags		图片
// @Param		page	query		page.ImagePage	true	"查询条件"
// @Success	200		object		e.Error
// @Router		/api/v1/image/page [get]
func ImageGetPage(ctx *gin.Context) {
	var p page.ImagePage
	var entity ds.Image
	err := ctx.ShouldBind(&p)
	e.PanicIfErrTipMsg(err, "参数错误")
	entity.GetPage(&p)
	app.Response(ctx, e.SuccessData(p))
}

// @Summary	单条数据
// @Tags		图片
// @Param		id		path		int		true	"ID"
// @Success	200		object		e.Error
// @Router		/api/v1/image/info/{id} [get]
func ImageGetById(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	e.PanicIfErr(err)
	m := ds.Image{Id: id}.GetById()
	app.Response(ctx, e.SuccessData(m))
}

// @Summary	删除图片
// @Tags		图片
// @Param		id		path		int		true	"ID"
// @Security	ApiKeyAuth
// @Router		/api/v1/image/del/{id} [post]
func ImageDel(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	e.PanicIfErr(err)

	img := ds.Image{Id: id}.GetById()
	// 删除文件（从 url 中解析 userId 和文件名）
	prefix := strings.TrimRight(setting.Config.App.FilePath.Image, "/") + "/"
	rel := strings.TrimPrefix(img.Url, prefix)
	parts := strings.SplitN(rel, "/", 2)
	if len(parts) == 2 {
		uid := parts[0]
		filename := parts[1]
		service.DeleteImageFile(uid, filename)
	}
	ds.Image{Id: id}.Delete()
	app.Response(ctx, e.Success)
}

// @Summary	用户图片列表（仅当前用户）
// @Tags		图片
// @Param		bizType	query		int		true	"业务类型"
// @Param		bizSubType	query		int		true	"业务子类型"
// @Security	ApiKeyAuth
// @Router		/api/v1/image/list [get]
func ImageList(ctx *contextx.AppContext) {
	// bizType 必填，且不能为 0
	s1 := ctx.Gin.Query("bizType")
	e.PanicIf(s1 == "", "参数异常：bizType")
	v1, err := strconv.ParseUint(s1, 10, 8)
	e.PanicIfErr(err)
	bizType := uint8(v1)
	e.PanicIf(bizType == 0, "参数异常：bizType")
	// bizSubType 可选，默认为 0
	v2, err := strconv.ParseUint(ctx.Gin.Query("bizSubType"), 10, 8)
	if (err) != nil {
		v2 = 0
	}
	bizSubType := uint8(v2)

	// addr 必填
	addr := ctx.Gin.Query("addr")
	e.PanicIf(addr == "", "addr 不能为空")

	items := ds.Image{}.GetList(bizType, bizSubType, addr)
	app.Response(ctx.Gin, e.SuccessData(items))
}

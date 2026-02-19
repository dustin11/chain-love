package ds_api

import (
	"chain-love/domain/active"
	"chain-love/domain/ds"
	"chain-love/domain/ds/enum"
	"chain-love/pkg/app"
	"chain-love/pkg/app/contextx"
	"chain-love/pkg/e"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

// @Summary	保存/更新便签
// @Tags		便签
// @Security	ApiKeyAuth
// @Router		/api/v1/note/save [post]
func NoteSave(ctx *contextx.AppContext) {
	var payload struct {
		Items []struct {
			Id    uint64 `json:"id"`
			Text  string `json:"text"`
			Style struct {
				FontSize   string     `json:"fontSize"` // 统一字符
				FontFamily string     `json:"fontFamily"`
				BgColor    string     `json:"bgColor"`
				Pos        [4]float64 `json:"pos"`
			} `json:"style"`
		} `json:"items"`
	}
	e.PanicIfErr(ctx.Gin.ShouldBindJSON(&payload))
	e.PanicIf(len(payload.Items) == 0, "items不能为空")

	var (
		toInsert    []ds.Note
		insertLocal []int // parallel slice of client local ids (it.Id)
		createdResp []map[string]interface{}
	)

	for _, it := range payload.Items {
		// validate Text
		txt := strings.TrimSpace(it.Text)
		if txt == "" {
			continue
		}
		e.PanicIf(utf8.RuneCountInString(txt) >= 800, "便签内容长度越出限制")

		styleB, _ := json.Marshal(it.Style)

		// 统一构造 Note（含 id/text/style），插入时把 id 清零
		n := ds.Note{
			Id:    it.Id,
			Text:  txt,
			Style: string(styleB),
		}
		if it.Id <= 10 {
			// 新建：不要传入 id，让 DB 自增
			n.Init(ctx.User)
			toInsert = append(toInsert, n)
			insertLocal = append(insertLocal, int(it.Id))
		} else {
			// 更新：只更新 text & style
			e.PanicIfErr(n.Update(ctx.User.Id))
		}
	}
	// batch insert new notes and return created items with localId
	if len(toInsert) > 0 {
		e.PanicIfErr(ds.AddBatchNote(toInsert))
		for i, ni := range toInsert {
			createdResp = append(createdResp, map[string]interface{}{
				"localId": insertLocal[i],
				"id":      ni.Id,
				"text":    ni.Text,
				"style":   ni.Style,
			})
		}
	}

	app.Response(ctx.Gin, e.SuccessData(createdResp))
}

// @Summary	便签列表
// @Tags		便签
// @Router		/api/v1/note/list [get]
func NoteList(c *gin.Context) {
	addr := c.Query("addr")
	e.PanicIf(addr == "", "addr 不能为空")

	items := ds.Note{}.GetList(addr)
	app.Response(c, e.SuccessData(items))
}

// @Summary	删除便签
// @Tags		便签
// @Security	ApiKeyAuth
// @Router		/api/v1/note/del [post]
func NoteDel(ctx *contextx.AppContext) {
	var req struct {
		Id uint64 `json:"id"`
	}
	e.PanicIfErr(ctx.Gin.ShouldBindJSON(&req))
	e.PanicIf(req.Id == 0, "id不能为空")

	n := ds.Note{Id: req.Id}.GetById()
	e.PanicIf(n.Id == 0 || n.CreatedBy != ctx.User.Id, "无法删除")
	// 删除便签记录
	ds.Note{Id: req.Id}.Delete(ctx.User.Id)
	// 删除关联的 like 数据（按 biz_type + data_id）
	e.PanicIfErr((&active.Like{Id: req.Id, BizType: uint8(enum.DeskSidingNote)}).DeleteByDataId())
	app.Response(ctx.Gin, e.Success)
}

package ds

import (
	"encoding/json"
	"log"

	"chain-love/domain"
	"chain-love/domain/ds/vo"
	"chain-love/pkg/app/security"
)

type Note struct {
	Id      uint64 `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Text    string `json:"text,omitempty" form:"text" gorm:"type:text;comment:便签文本"`
	Style   string `json:"style,omitempty" form:"style" gorm:"type:json;comment:样式数据"`
	Creator string `json:"creator,omitempty" form:"creator" gorm:"type:varchar(64);comment:创建者钱包地址"`

	domain.CreatInfo
}

func (Note) TableName() string {
	return "ds_note"
}

func (m *Note) Init(user *security.JwtUser) *Note {
	m.Id = 0 // 确保新建时 ID 为 0
	m.Creator = user.Addr
	m.CreatedBy = user.Id
	return m
}

func (m *Note) Add() error {
	return domain.Db.Create(m).Error
}

// AddBatch 批量插入（GORM 会回填自增主键）
func AddBatchNote(notes []Note) error {
	return domain.Db.Create(&notes).Error
}

func (m *Note) Update(userId uint64) error {
	// 仅更新 text 与 style 字段，基于 id + created_by 权限校验
	result := domain.Db.Model(&Note{}).
		Where("id = ? AND created_by = ?", m.Id, userId).
		Updates(map[string]interface{}{"text": m.Text, "style": m.Style})

	if result.Error != nil {
		log.Println("Note Update Error", result.Error)
		return result.Error
	}
	// if result.RowsAffected == 0 {
	// 	return errors.New("无权修改或记录不存在")
	// }
	return nil
}

func (m Note) Delete(userId uint64) {
	domain.Db.Where("id = ? AND created_by = ?", m.Id, userId).Delete(&Note{})
}

func (m Note) GetById() Note {
	var note Note
	domain.Db.First(&note, m.Id)
	return note
}

// ToVO 将 Note 转换为前端需要的 VO（解析 Style JSON）
func (m Note) ToVO() vo.NoteVO {
	var styleObj *vo.NoteStyleVO
	if m.Style != "" {
		var s vo.NoteStyleVO
		if err := json.Unmarshal([]byte(m.Style), &s); err == nil {
			styleObj = &s
		}
	}
	return vo.NoteVO{
		Id:    m.Id,
		Text:  m.Text,
		Style: styleObj,
	}
}

// GetList 返回前端需要的 VO（按 creator 查询）
func (m Note) GetList(addr string) []vo.NoteVO {
	var list []Note
	db := domain.Db.Where("creator = ?", addr)
	db.Find(&list)

	res := make([]vo.NoteVO, 0, len(list))
	for _, it := range list {
		res = append(res, it.ToVO())
	}
	return res
}

package ds

import (
	"errors"
	"log"

	"chain-love/domain"
	"chain-love/domain/ds/page"
	"chain-love/pkg/app/security"
)

type Image struct {
	Id         uint64 `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Url        string `json:"url" form:"url" gorm:"type:varchar(255);not null;comment:图片路径"`
	BizType    uint8  `json:"bizType,omitempty" form:"bizType" gorm:"type:tinyint;comment:业务类型"`
	BizSubType uint8  `json:"bizSubType,omitempty" form:"bizSubType" gorm:"type:tinyint;comment:业务子类"`
	Style      string `json:"style,omitempty" form:"style" gorm:"type:json;comment:样式数据"`
	Creator    string `json:"creator,omitempty" form:"creator" gorm:"type:varchar(64);comment:创建者钱包地址"`

	domain.CreatInfo
}

func (Image) TableName() string {
	return "ds_image"
}

func (m *Image) Init(user *security.JwtUser) *Image {
	m.Creator = user.Addr
	m.CreatedBy = user.Id
	return m
}

func (m *Image) Add() error {
	return domain.Db.Create(m).Error
}

// AddBatch 批量插入（GORM 会回填自增主键）
func AddBatch(images []Image) error {
	return domain.Db.Create(&images).Error
}

// Update 常规更新（基于 created_by 权限校验）
func (m *Image) Update(userId uint64) error {
	result := domain.Db.Where("id = ? AND created_by = ?", m.Id, userId).Updates(m)
	if result.Error != nil {
		log.Println("Image Update Error", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("无权修改或记录不存在")
	}
	return nil
}

// UpdateStyle 仅更新 style 字段（基于 id + created_by 校验权限）
func UpdateStyle(id uint64, userId uint64, style string) error {
	result := domain.Db.Model(&Image{}).
		Where("id = ? AND created_by = ?", id, userId).
		Update("style", style)

	if result.Error != nil {
		log.Println("Image UpdateStyle Error", result.Error)
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("无权修改或记录不存在")
	}
	return nil
}

func (m Image) Delete() {
	domain.Db.Where("id = ?", m.Id).Delete(&Image{})
}

func (m Image) GetById() Image {
	var img Image
	domain.Db.First(&img, m.Id)
	return img
}

func (m Image) GetPage(page *page.ImagePage) {
	var list []*Image
	db := domain.Db

	if page.Id != 0 {
		db = db.Where("id = ?", page.Id)
	}
	// BizType/ BizSubType are enums (numeric), check non-zero
	if page.BizType != 0 {
		db = db.Where("biz_type = ?", page.BizType)
	}
	if page.BizSubType != 0 {
		db = db.Where("biz_sub_type = ?", page.BizSubType)
	}
	if page.Url != "" {
		db = db.Where("url like ?", "%"+page.Url+"%")
	}
	if page.Creator != "" {
		db = db.Where("creator = ?", page.Creator)
	}

	page.SetModel(&Image{}).
		SetRecords(&list).
		QueryPage(db)
}

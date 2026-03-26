package dev

import (
	"errors"

	"senspace/domain"
	"senspace/pkg/app/security"
	"senspace/pkg/util"
)

type Plugin struct {
	Id          int64  `json:"id,string" form:"id" gorm:"primaryKey;autoIncrement:false;comment:ID"`
	Name        string `json:"name" form:"name" gorm:"type:varchar(255);not null;comment:名称"`
	Description string `json:"description" form:"description" gorm:"type:varchar(500);comment:描述"`
	Version     string `json:"version" form:"version" gorm:"type:varchar(50);comment:版本"`
	Author      string `json:"author" form:"author" gorm:"type:varchar(100);comment:作者"`
	domain.CreatInfo
}

func (Plugin) TableName() string {
	return "dev_plugin"
}

func (m *Plugin) Init(user *security.JwtUser) *Plugin {
	m.Id = util.IdWorker.Generate().Int64()
	m.Author = user.Nickname
	m.CreatedBy = user.Id
	return m
}

func (m *Plugin) Add() error {
	return domain.Db.Create(m).Error
}

func (m *Plugin) Update(userId uint64) error {
	result := domain.Db.Where("id = ? AND created_by = ?", m.Id, userId).Updates(m)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("无权修改或记录不存在")
	}
	return nil
}

func (m Plugin) Delete() {
	domain.Db.Where("id = ?", m.Id).Delete(&Plugin{})
}

func (m Plugin) GetById() Plugin {
	var plugin Plugin
	domain.Db.First(&plugin, m.Id)
	return plugin
}

func (Plugin) List() []*Plugin {
	var list []*Plugin
	domain.Db.Find(&list)
	return list
}

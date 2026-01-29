package sys

import (
	"chain-love/domain"
	page2 "chain-love/domain/sys/page"
	"log"
)

type Role struct {
	Id     uint64 `gorm:"primaryKey;autoIncrement" json:"id" form:"id"`
	Name   string `json:"name" form:"name" example:"名称"`
	Remark string `json:"remark" form:"remark" example:"注备"`
	domain.CreatInfo

	MenuIds []int64 `json:"menuIds" gorm:"-"`
}

func (role Role) TableName() string {
	return "sys_role"
}

func (entity Role) Add() bool {
	create := domain.Db.Create(&entity)
	return create.Error != nil
}

func (entity *Role) Update() error {
	//e :=
	//	model.Db.Where("id = ?", entity.Id).
	//		Update(entity).Error
	e := domain.Db.Model(&entity).Updates(&entity).Error
	if e != nil {
		log.Panicln("发生了错误", e.Error())
	}
	return e
}

func (m Role) Delete() {
	domain.Db.Where("id = ?", m.Id).Delete(Role{})
}

func (entity Role) QueryById() Role {
	domain.Db.First(&entity, entity.Id)
	return entity
}

func (entity Role) GetPage(page *page2.RolePage) {
	var list []*Role
	db := domain.Db
	if page.Name != "" {
		//db = db.Where("name = ?", page.Name)
		//db = db.Where(&Menu{Name:page.Name})
		db = db.Where("name like ?", "%"+page.Name+"%")
	}
	//page查询封装
	page.SetModel(&Role{}).
		SetRecords(&list).
		QueryPage(db)
}

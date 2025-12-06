package ds

import (
	"chain-love/domain"
	"chain-love/domain/ds/enum"
	page2 "chain-love/domain/sys/page"
	"chain-love/pkg/util"
	"log"
)

type Room struct {
	Id          int64                `json:"id" gorm:"primary_key;" form:"id"`
	HouseId     string               `json:"houseId" form:"houseId" gorm:"not null"  example:"房源id"`
	No          string               `json:"No" form:"No" gorm:"size:15"  example:"编号"`
	Orientation enum.OrientationType `json:"orientation" form:"orientation" desc:"朝向"`
	Furniture   enum.FurnitureType   `json:"furniture" form:"furniture" desc:"房屋设施"`
	Imgs        string               `json:"imgs" form:"imgs" gorm:"size:512" desc:"图片，分隔"`
	Remark      string               `json:"remark" form:"remark" gorm:"size:4800" example:"备注"`
	RentState   enum.RentState       `json:"rentState" form:"rentState" example:"出租状态"`
	RentId      int64                `json:"rentId" example:"0" desc:"对应出租表id, 出租结束置空"`
	State       enum.RowState        `json:"state" form:"state" gorm:"not null;default:1" desc:"数据状态"`
	domain.AreaModel
}

func (m Room) TableName() string {
	return "ds_room"
}

func (m Room) getId() {

}

func (m Room) Add() bool {
	if m.Id == 0 {
		m.Id = util.IdWorker.Generate().Int64()
	}
	create := domain.Db.Create(&m)
	return create.Error != nil
}

func (entity *Room) Update() error {
	//e :=
	//	model.Db.Where("id = ?", entity.Id).
	//		Update(entity).Error
	e := domain.Db.Model(&entity).Updates(&entity).Error
	if e != nil {
		log.Panicln("发生了错误", e.Error())
	}
	return e
}

func (m Room) Delete() {
	domain.Db.Where("id = ?", m.Id).Delete(Room{})
}

func (entity Room) QueryById() Room {
	domain.Db.First(&entity, entity.Id)
	return entity
}

func (entity Room) GetPage(page *page2.RolePage) {
	var list []*Room
	db := domain.Db
	if page.Name != "" {
		//db = db.Where("name = ?", page.Name)
		//db = db.Where(&Menu{Name:page.Name})
		db = db.Where("name like ?", "%"+page.Name+"%")
	}
	//page查询封装
	page.SetModel(&Room{}).
		SetRecords(&list).
		QueryPage(db)
}

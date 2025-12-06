package ds

import (
	"chain-love/domain"
	"chain-love/domain/ds/enum"
	"chain-love/domain/ds/page"
	"chain-love/pkg/e"
	"chain-love/pkg/util"
	"log"
)

type House struct {
	Id            uint64               `json:"id" gorm:"primary_key" form:"id"`
	Village       string               `json:"village" form:"village" gorm:"not null;size:27" example:"小区名"`
	BedroomNum    int                  `json:"bedroomNum" form:"bedroomNum" gorm:"not null" example:"卧室数量"`
	LivingRoomNum int                  `json:"livingRoomNum" form:"livingRoomNum" desc:"客厅数量"`
	BathroomNum   int                  `json:"bathroomNum" form:"bathroomNum" desc:"卫生间数量"`
	Square        float64              `json:"square,string" form:"square" gorm:"not null" desc:"平米"`
	Floor         int                  `json:"floor" form:"floor" desc:"层数"`
	TotalFloor    int                  `json:"totalFloor" form:"totalFloor" desc:"总层数"`
	HasLift       bool                 `json:"hasLift" form:"hasLift" gorm:"not null" desc:"有无电梯"`
	HasParking    bool                 `json:"hasParking" form:"hasParking" gorm:"not null" desc:"有无车位"`
	Decoration    enum.DecorationType  `json:"decoration" form:"decoration" desc:"装修情况"`
	Furniture     enum.FurnitureType   `json:"furniture" form:"furniture" desc:"房屋设施"`
	Orientation   enum.OrientationType `json:"orientation" form:"orientation" desc:"朝向"`
	Imgs          string               `json:"Imgs" form:"Imgs" desc:"图片，分隔"`
	Remark        string               `json:"remark" form:"remark" gorm:"size:4800" example:"备注"`
	Type          enum.HouseType       `json:"type" form:"type" gorm:"not null" desc:"房源类型"`
	RentState     enum.RentState       `json:"rentState" form:"rentState" gorm:"not null;default:1" desc:"出租状态"`
	RentId        int64                `json:"rentId" form:"rentId" gorm:"force" example:"0" desc:"对应出租表id, 出租结束置空"`
	State         enum.RowState        `json:"state" form:"state" gorm:"not null;default:1" desc:"数据状态"`
	domain.AreaModel
}

func (m House) TableName() string {
	return "ds_house"
}

func (m House) Valid() House {
	e.PanicIf(m.Village == "", "小区名称不能为空！")
	e.PanicIf(m.BedroomNum <= 0, "卧室数量不能为零！")
	e.PanicIf(m.Orientation == 0, "请选择朝向！")
	e.PanicIf(m.Imgs == "", "请上传图片！")
	e.PanicIf(m.Type == 0, "房源类型不能为空！")

	return m
}

func (m House) Init() House {
	if m.Id == 0 { //添加
		m.Id = uint64(util.IdWorker.Generate().Int64())
	}
	return m
}

func (m *House) Save() {
	if m.State != enum.Row_State_Draft {
		m.Valid()
	}
	if m.Id == 0 {
		m.Init().Add()
	} else {
		m.Update()
	}
}

func (entity House) Add() bool {
	create := domain.Db.Create(&entity)
	return create.Error != nil
}

func (entity House) Update() error {
	e := domain.Db.Model(entity).Updates(entity).Error
	if e != nil {
		log.Panicln("发生了错误", e.Error())
	}
	return e
}

func (m House) Delete() {
	domain.Db.Where("id = ?", m.Id).Delete(House{})
}

func (entity House) QueryById() House {
	domain.Db.First(&entity, entity.Id)
	return entity
}

func (entity House) GetPage(page *page.HousePage) {
	var list []*House
	db := domain.Db
	if page.Name != "" {
		db = db.Where("name like ?", "%"+page.Name+"%")
	}
	//page查询封装
	page.SetModel(&House{}).
		SetRecords(&list).
		QueryPage(db)
}

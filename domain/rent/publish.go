package rent

import (
	"chain-love/domain"
	"chain-love/domain/ds/enum"
)

// 出租表
type Publish struct {
	Id          int64            `gorm:"primary_key" json:"id" form:"id"`
	SourceId    int64            `gorm:"not null" json:"roomId" form:"roomId" desc:"house id或room id或... id"`
	SourceType  enum.HouseType   `gorm:"not null" json:"sourceType" form:"sourceType" desc:"房源类型"`
	FromDate    domain.Date      `gorm:"force;type:date" json:"fromDate" form:"fromDate" desc:"出租开始时间"`
	Rent        float32          `gorm:"not null" json:"rent,string" form:"rent" desc:"租金"`
	PayType     enum.RentPayType `gorm:"not null" json:"payType,string" form:"payType" desc:"付款方式"`
	RentInclude enum.RentInclude `json:"rentInclude,string" form:"rentInclude" desc:"租金包含"`
	Remark      string           `gorm:"size:4800" json:"Remark,string" form:"Remark" desc:"房源描述 可以介绍一下房源亮点，交通、周边环境，可以入住的时间和对租客的要求等，详细的描述会大大增加快速出租的机会！"`
	State       enum.DataState   `gorm:"not null" json:"state" form:"state" desc:"状态"`
}

func (m Publish) TableName() string {
	return "ren_publish"
}

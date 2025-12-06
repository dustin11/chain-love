package rent

import (
	"chain-love/domain"
	"chain-love/domain/ds/enum"
)

// 出租表
type Rent struct {
	Id        int64         `json:"id" gorm:"primary_key;" form:"id"`
	SourceId  int64         `gorm:"not null" json:"sourceId" form:"sourceId" desc:"house或room或... id"`
	UserId    int64         `gorm:"not null" json:"userId" form:"userId" example:"用户id"`
	Type      enum.RentType `gorm:"force" json:"type" form:"type" example:"0" desc:"租类型"`
	FromDate  domain.Date   `gorm:"not null;type:date" json:"fromDate" form:"fromDate" desc:"租期开始时间"`
	EndDate   domain.Date   `gorm:"not null;type:date" json:"endDate" form:"endDate" desc:"租期结束时间"`
	Remark    string        `gorm:"size:512" json:"remark" form:"remark" example:"备注"`
	CreatedOn domain.Time   `gorm:"type:datetime" json:"createdOn"`
}

func (m Rent) TableName() string {
	return "ren_rent"
}

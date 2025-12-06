package app

import (
	"gorm.io/gorm"
)

type AreaPagination struct {
	//Ccid uint64 `json:"ccid" form:"ccid" desc:"记录总数" example:"100"`
	Province string `json:"province" form:"province" gorm:"size:27" example:"省"`
	City     string `json:"city" form:"city" example:"市"`
	County   string `json:"county" form:"county" example:"县"`
	Town     string `json:"town" form:"town" example:"镇"`
	Pagination
}

func (page *AreaPagination) SetModel(model interface{}) *AreaPagination {
	page.Model = model
	return page
}

func (page *AreaPagination) SetRecords(records interface{}) *AreaPagination {
	page.Records = records
	return page
}

func (page *AreaPagination) QueryPage(db *gorm.DB) {
	//e.PanicIf(page.Ccid==0,"公司id不能为空")
	//db = db.Where("ccid = ?", G.User.Ccid)
	db = db.Where("city = ?", 1)
	page.Pagination.QueryPage(db)
}

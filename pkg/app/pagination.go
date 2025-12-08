package app

import (
	"chain-love/pkg/e"

	"gorm.io/gorm"
)

const INT_MAX = int(^uint(0) >> 1)

type IPagination interface {
	SetModel(model interface{})
	SetRecords(records interface{})
	getSize() uint
	getCurrent() uint
	QueryPage(db *gorm.DB)
}

type Pagination struct {
	Total     uint64      `json:"total" form:"total" desc:"记录总数" example:"100"`
	Size      uint        `json:"size" form:"size" desc:"每页条数" example:"10"`
	Current   uint        `json:"current" form:"current" desc:"当前页数" example:"1"`
	UnPage    bool        `json:"unPage" form:"unPage" desc:"不分页" example:"false"`
	OrderAsc  string      `json:"orderAsc" form:"orderAsc" desc:"升序" example:""`
	OrderDesc string      `json:"orderDesc" form:"orderDesc" desc:"降序" example:""`
	Records   interface{} `json:"records" swaggertype:"string"`
	Model     interface{} `json:"-" swaggertype:"string"`
}

func (page *Pagination) SetModel(model interface{}) *Pagination {
	page.Model = model
	return page
}

func (page *Pagination) SetRecords(records interface{}) *Pagination {
	page.Records = records
	return page
}

func (page *Pagination) getSize() uint {
	if page.UnPage {
		return uint(INT_MAX)
	}
	if page.Size <= 0 {
		return 10
	}
	return page.Size
}

func (page *Pagination) getCurrent() uint {
	if page.Current <= 0 {
		return 1
	}
	return page.Current
}

func (page *Pagination) QueryPage(db *gorm.DB) {
	if page.Total <= 0 && !page.UnPage {
		var total int64
		db.Model(page.Model).Count(&total) //总行数
		page.Total = uint64(total)
	} else {
		page.Total = uint64(INT_MAX)
	}
	if page.OrderAsc != "" {
		db = db.Order(page.OrderAsc + " asc")
	} else if page.OrderDesc != "" {
		db = db.Order(page.OrderDesc + " desc")
	}
	err :=
		db.Limit(int(page.getSize())).Offset(int((page.getCurrent() - 1) * page.getSize())).
			Find(page.Records).Error

	e.PanicIfErr(err)
}

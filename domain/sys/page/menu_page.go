package page

import "chain-love/pkg/app"

type MenuPage struct {
	Name   string `json:"name" form:"name" example:"名称"`
	Type   []int  `json:"type" form:"type" example:"0" desc:"类型"`
	UserId uint64 `json:"userId" form:"userId" example:"0" desc:"用户id"`
	app.Pagination
}

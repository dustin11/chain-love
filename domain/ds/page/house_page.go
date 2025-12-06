package page

import "chain-love/pkg/app"

type HousePage struct {
	Name string `json:"name" form:"name" example:"名称"`
	app.AreaPagination
}

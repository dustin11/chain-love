package page

import "senspace/pkg/app"

type HousePage struct {
	Name string `json:"name" form:"name" example:"名称"`
	app.AreaPagination
}

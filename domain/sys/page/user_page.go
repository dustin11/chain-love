package page

import "chain-love/pkg/app"

type UserPage struct {
	Username string `json:"username" form:"username" example:"用户名"`
	app.AreaPagination
}

package page

import "chain-love/pkg/app"

type UserPage struct {
	Nickname string `json:"nickname" form:"nickname" example:"昵称"`
	app.AreaPagination
}

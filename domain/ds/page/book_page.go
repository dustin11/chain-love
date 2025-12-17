package page

import "chain-love/pkg/app"

type BookPage struct {
	Title    string `json:"title" form:"title"`
	Author   string `json:"author" form:"author"`
	Language string `json:"language" form:"language"`
	CateId   int    `json:"cateId" form:"cateId"`
	app.Pagination
}

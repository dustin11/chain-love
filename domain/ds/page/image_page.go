package page

import (
	"chain-love/domain/ds/enum"
	"chain-love/pkg/app"
)

type ImagePage struct {
	Id         uint64          `json:"id" form:"id"`
	BizType    enum.ImgBizType `json:"bizType" form:"bizType"`
	BizSubType enum.ImgBizType `json:"bizSubType" form:"bizSubType"`
	Url        string          `json:"url" form:"url"`
	Creator    string          `json:"creator" form:"creator"`
	CreatedBy  uint64          `json:"createdBy" form:"createdBy"`
	app.Pagination
}

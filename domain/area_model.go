package domain

type AreaModel struct {
	Country  string `json:"country" form:"country" gorm:"size:52" example:"国家"`
	Province string `json:"province" form:"province" gorm:"size:52" example:"省"`
	City     string `json:"city" form:"city" gorm:"size:52" example:"市"`
	Area     string `json:"area" form:"area" gorm:"size:52" example:"区"`
}

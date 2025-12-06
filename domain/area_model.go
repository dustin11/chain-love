package domain

type AreaModel struct {
	Model
	Province string `json:"province" form:"province" gorm:"size:27" example:"省"`
	City     string `json:"city" form:"city" gorm:"size:27" example:"市"`
	County   string `json:"county" form:"county" gorm:"size:27" example:"县"`
	Town     string `json:"town" form:"town" gorm:"size:27" example:"镇"`
}

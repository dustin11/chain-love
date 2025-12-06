package domain

type MiniUser struct {
	UserId   int64  `json:"userId" form:"userId"`
	NickName string `json:"nickName" form:"nickName" example:"昵称"`
	Avatar   string `json:"avatar" form:"avatar" example:"头像"`
}

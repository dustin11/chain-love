package models

import (
	"chain-love/domain"
	"chain-love/domain/sys"
)

// 前端展示用的 User DTO（只包含需要暴露的字段）
type UserDTO struct {
	ID       uint64 `json:"id"`
	Addr     string `json:"addr"`
	Nickname string `json:"nickname,omitempty"`
	// Email       string `json:"email,omitempty"`
	// Mobile      string `json:"mobile,omitempty"`
	Avatar      string `json:"avatar,omitempty"`
	State       byte   `json:"state"`
	AccountPart byte   `json:"accountPart"`
	domain.AreaModel
}

// 将领域实体转换为前端 DTO（显式映射，便于后续演进）
func NewUserDTO(u *sys.User) UserDTO {
	if u == nil {
		return UserDTO{}
	}
	return UserDTO{
		ID:       u.Id,
		Addr:     u.Addr,
		Nickname: u.Nickname,
		// Email:       u.Email,
		// Mobile:      u.Mobile,
		Avatar:      u.Avatar,
		State:       u.State,
		AccountPart: u.AccountPart,
		AreaModel:   u.AreaModel,
	}
}

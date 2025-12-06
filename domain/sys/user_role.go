package sys

import (
	"bytes"
	"chain-love/domain"
	"fmt"
)

type UserRole struct {
	Id     uint64 `gorm:"primary_key;AUTO_INCREMENT" json:"id" form:"id"`
	UserId uint64 `json:"menuId" form:"userId" desc:"用户id"`
	RoleId uint64 `json:"roleId" form:"roleId" desc:"角色id"`
}

func (entity UserRole) TableName() string {
	return "sys_user_role"
}

func (entity UserRole) Add(roleIds []int64) error {
	var buffer bytes.Buffer
	sql := fmt.Sprintf("insert into `%s` (`user_id`,`role_id`) values", UserRole{}.TableName())
	if _, err := buffer.WriteString(sql); err != nil {
		return err
	}
	for i, e := range roleIds {
		if i == len(roleIds)-1 {
			buffer.WriteString(fmt.Sprintf("('%d','%d');", entity.UserId, e))
		} else {
			buffer.WriteString(fmt.Sprintf("('%d','%d'),", entity.UserId, e))
		}
	}
	return domain.Db.Exec(buffer.String()).Error
}

func (m UserRole) DeleteByUserId() {
	domain.Db.Where("user_id = ?", m.UserId).Delete(UserRole{})
}

func (m UserRole) GetByUserId() []UserRole {
	var list []UserRole
	domain.Db.Where("user_id = ?", m.UserId).Find(&list)
	return list
}

func (m UserRole) GetByRoleId() []UserRole {
	var list []UserRole
	domain.Db.Where("role_id = ?", m.RoleId).Find(&list)
	return list
}

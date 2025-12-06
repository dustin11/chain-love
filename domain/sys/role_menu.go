package sys

import (
	"bytes"
	"chain-love/domain"
	"chain-love/pkg/logging"
	"fmt"
)

type RoleMenu struct {
	Id     uint64 `gorm:"primary_key;AUTO_INCREMENT" json:"id" form:"id"`
	RoleId uint64 `json:"roleId" form:"roleId" desc:"角色id"`
	MenuId int64  `json:"menuId" form:"menuId" desc:"菜单id"`
}

func (entity RoleMenu) TableName() string {
	return "sys_role_menu"
}

func (entity RoleMenu) Add(roleId uint64, menuIds []int64) error {
	var buffer bytes.Buffer
	sql := fmt.Sprintf("insert into `%s` (`role_id`,`menu_id`) values", RoleMenu{}.TableName())
	if _, err := buffer.WriteString(sql); err != nil {
		return err
	}
	for i, e := range menuIds {
		logging.Info(e)
		if i == len(menuIds)-1 {
			buffer.WriteString(fmt.Sprintf("('%d','%d');", roleId, e))
		} else {
			buffer.WriteString(fmt.Sprintf("('%d','%d'),", roleId, e))
		}
	}
	return domain.Db.Exec(buffer.String()).Error
}

func (entity RoleMenu) Delete(roleId uint64) {
	domain.Db.Where("role_id = ?", roleId).Delete(RoleMenu{})
}

func (m RoleMenu) GetRoleMenu() []RoleMenu {
	var list []RoleMenu
	domain.Db.Where("role_id = ?", m.RoleId).Find(&list)
	return list
}

package sys_service

import (
	"chain-love/domain/sys"
)

func SaveRole(m sys.Role) {
	if m.Id == 0 {
		addRole(m)
	} else {
		updateRole(m)
	}
}

func addRole(m sys.Role) {
	m.Add()
	//添加角色菜单
	sys.RoleMenu{}.Add(m.Id, m.MenuIds)
}

func updateRole(m sys.Role) {
	m.Update()
	//删除角色菜单
	sys.RoleMenu{}.Delete(m.Id)
	//添加角色菜单
	sys.RoleMenu{}.Add(m.Id, m.MenuIds)
}

func GetRoleIds(uid uint64) []int64 {
	userRole := sys.UserRole{UserId: uid}.GetByUserId()
	ids := []int64{}
	for _, i := range userRole {
		ids = append(ids, int64(i.RoleId))
	}
	return ids
}

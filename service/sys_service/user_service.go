package sys_service

import (
	"chain-love/domain/sys"
	"chain-love/pkg/e"
)

func UserSave(m *sys.User) {
	if m.Id == 0 {
		UserAdd(m)
	} else {
		userUpdate(m)
	}
}

func UserAdd(m *sys.User) {
	//添加用户
	m.Valid().Init().Add()
	//添加角色
	if len(m.RoleIds) > 0 {
		sys.UserRole{UserId: m.Id}.Add(m.RoleIds)
	}
}

func userUpdate(m *sys.User) {
	user := m.GetByUsernameNotId()
	e.PanicIf(user.Id > 0, "用户名已被占用！")
	m.Update()
	//处理角色
	userRole := sys.UserRole{UserId: m.Id}
	userRole.DeleteByUserId()
	if len(m.RoleIds) > 0 {
		userRole.Add(m.RoleIds)
	}

}

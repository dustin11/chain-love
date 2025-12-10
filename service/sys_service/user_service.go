package sys_service

import (
	"chain-love/domain/sys"
	"chain-love/pkg/e"
)

func UserSave(m *sys.User) {
	e.PanicIf(m.Id == 0, "用户不存在！")
	userUpdate(m)
}

func userUpdate(m *sys.User) {
	user := m.GetByAddrNotId()
	e.PanicIf(user.Id > 0, "用户名已被占用！")
	m.Update()
	//处理角色
	userRole := sys.UserRole{UserId: m.Id}
	userRole.DeleteByUserId()
	// if len(m.RoleIds) > 0 {
	// 	userRole.Add(m.RoleIds)
	// }

}

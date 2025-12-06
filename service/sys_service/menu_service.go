package sys_service

import (
	"chain-love/domain/sys"
	page2 "chain-love/domain/sys/page"
	"chain-love/pkg/app"
	"chain-love/pkg/e"
	"chain-love/pkg/setting/consts"
	"strings"
)

//func GetMenuPage(page page.MenuPage) app.Pagination {
//	var list []*model.Menu
//	db := page
//	if page.Name != "" {
//		//db = db.Where("name = ?", page.Name)
//		//db = db.Where(&Menu{Name:page.Name})12
//		db = db.Where("name like ?", "%"+page.Name+"%")
//	}
//	page.GetPageQuery(db)
//}

// 获取菜单 + 用户权限
func GetMenuWithPerms(userId uint64) map[string]interface{} {
	var m = &sys.Menu{}
	var perms []string
	var menus *[]*sys.Menu
	if userId == consts.ADMIN {
		page := &page2.MenuPage{Pagination: app.Pagination{UnPage: true, OrderAsc: "order_num"}} //, Type:[]int{0,1}
		m.GetPage(page)
		menus, _ = page.Records.(*[]*sys.Menu)
	} else {
		ms := m.GetUserMenu(userId)
		menus = &ms
	}
	perms = getPerms(menus)
	menuTree := GetMenuTree(menus)
	return map[string]interface{}{"menuList": menuTree, "perms": perms}
}

func getPerms(menus *[]*sys.Menu) []string {
	var perms []string
	//用于去重
	permsMap := make(map[string]bool)
	for _, i := range *menus {
		if i.Perms == "" {
			continue
		}
		//处理逗号分隔
		for _, j := range strings.Split(i.Perms, ",") {
			permsMap[j] = true
		}
	}
	for k, _ := range permsMap {
		perms = append(perms, k)
	}
	return perms
}

// 获取菜单树
func GetMenuTree(list *[]*sys.Menu) []*sys.Menu {
	var root []*sys.Menu
	//添加root
	for _, i := range *list {
		if i.ParentId == 0 {
			root = append(root, i)
			//添加child
			getChild(i, list)
			//i.Children = append(i.Children, childs)
		}
	}

	return root
}

func getChild(m *sys.Menu, list *[]*sys.Menu) {
	for _, i := range *list {
		if i.ParentId == m.Id {
			m.Children = append(m.Children, i)
			//目录，递归获取
			if i.Type == 0 {
				getChild(i, list)
			}
		}
	}
}

func GetMenuIds(roleId uint64) []int64 {
	list := []int64{}
	roleMenus := sys.RoleMenu{RoleId: roleId}.GetRoleMenu()
	for _, i := range roleMenus {
		list = append(list, int64(i.MenuId))
	}
	return list
}

func SaveMenu(m sys.Menu) {
	if m.Id == 0 {
		m.Add()
	} else {
		e.PanicIf(m.ParentId == m.Id, "父级不能为自己")
		m.Update()
	}
}

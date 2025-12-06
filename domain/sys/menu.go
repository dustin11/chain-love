package sys

import (
	"chain-love/domain"
	page2 "chain-love/domain/sys/page"
	"chain-love/pkg/setting/consts"
)

type Menu struct {
	Id       uint64 `json:"id" gorm:"primary_key;AUTO_INCREMENT" form:"id"`
	Name     string `json:"name" form:"name" example:"名称"`
	ParentId uint64 `json:"parentId" form:"parentId" example:"父级" gorm:"force"`
	Url      string `json:"url" form:"url" example:"菜单URL"`
	Perms    string `json:"perms" form:"perms" example:"权限码"`
	Type     byte   `gorm:"force" json:"type" form:"type" example:"0"`
	Icon     string `json:"icon" form:"icon" example:"图标"`
	OrderNum int    `json:"orderNum" form:"orderNum" desc:"序号" example:"0"`
	domain.Model

	Children []*Menu `json:"children" gorm:"-"`
}

func (menu Menu) TableName() string {
	return "sys_menu"
}

func (menu Menu) Add() bool {
	create := domain.Db.Create(&menu)
	return create.Error != nil
}

func (menu *Menu) Update() {
	//e :=
	//	model.Db.Model(&menu).Where("id = ?", menu.Id).
	//		Update("name", menu.Name, "url", menu.Url, "type", menu.Type, "parentId", menu.ParentId).Error
	//零值不更新 gorm:"force"不起作用 使用save更新全部 坑！
	domain.Db.Model(&menu).Save(&menu)
	//if e != nil {
	//	log.Panicln("发生了错误", e.Error())
	//}
}

func (m Menu) Delete() {
	domain.Db.Where("id = ?", m.Id).Delete(Menu{})
}

func (menu Menu) QueryById() Menu {
	domain.Db.First(&menu, menu.Id)
	return menu
}

func (menu Menu) GetPage(page *page2.MenuPage) {
	var list []*Menu
	db := domain.Db
	if page.Name != "" {
		//db = db.Where("name = ?", page.Name)
		//db = db.Where(&Menu{Name:page.Name})
		db = db.Where("name like ?", "%"+page.Name+"%")
	}
	if page.Type != nil {
		db = db.Where("type in (?)", page.Type)
	}
	if page.UserId > 0 && page.UserId != consts.ADMIN {
		db = db.Where("id <> ? and parent_id <> ?", consts.MENU_PAGE_ID, consts.MENU_PAGE_ID)
	}
	//page查询封装
	page.SetModel(&Menu{}).
		SetRecords(&list).
		QueryPage(db)
	//db = page.GetPageQuery(db)
	//err := db.Find(&list).Error
	//page.Records = &list
	//e.PanicIfErr(err)
}

func (menu Menu) GetUserMenu(userId uint64) []*Menu {
	var perms []*Menu
	domain.Db.Raw("select m.* from sys_user_role ur "+
		"LEFT JOIN sys_role_menu rm on ur.role_id = rm.role_id "+
		"inner JOIN sys_menu m on rm.menu_id = m.id where ur.user_id = ? and m.id<> ? and m.parent_id<> ?", userId, consts.MENU_PAGE_ID, consts.MENU_PAGE_ID).
		Scan(&perms)

	return perms
}

//func (menu Menu) GetUserPerms(userId uint64) []sql.NullString {
//	var perms []sql.NullString
//	model.Db.Raw("select m.perms from sys_user_role ur "+
//		"LEFT JOIN sys_role_menu rm on ur.role_id = rm.role_id "+
//		"LEFT JOIN sys_menu m on rm.menu_id = m.id where ur.user_id = ?", userId).
//		Pluck("perms", &perms)
//
//	return perms
//}
//
//func (menu Menu) GetUserMenuId(userId uint64) []uint64 {
//	var menuIds []uint64
//	model.Db.Raw("select distinct rm.menu_id from sys_user_role ur "+
//					"LEFT JOIN sys_role_menu rm on ur.role_id = rm.role_id "+
//					"where ur.user_id = ?", userId).
//			Pluck("perms", &menuIds)
//	return menuIds
//}

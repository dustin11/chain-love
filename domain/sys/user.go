package sys

import (
	"chain-love/domain"
	page2 "chain-love/domain/sys/page"
	"chain-love/pkg/app"
	"chain-love/pkg/e"
	"chain-love/pkg/setting"
	"chain-love/pkg/util"

	// "database/sql"
	"log"
)

type User struct {
	Id uint64 `json:"id" gorm:"primary_key" form:"id"`
	// Username    string           `json:"username" form:"username" gorm:"not null;size:52" example:"邮箱或手机号"` //form:后不能有空格，否则接不到值，坑爹
	Addr     string `json:"addr" form:"addr" gorm:"not null;size:52"` //钱包地址
	Nickname string `json:"nickname" form:"nickname" example:"昵称"`
	Email    string `json:"email" form:"email" example:"邮箱"`
	Mobile   string `json:"mobile" form:"mobile" example:"手机号"`
	State    byte   `json:"state" example:"0"`
	// Salt     sql.NullString `json:"-" example:"盐值当头像"`
	Avatar string `json:"avatar" form:"avatar" example:"头像"`
	domain.AreaModel
	AccountPart byte `json:"AccountPart" gorm:"not null" ` //帐号完整度
	domain.Model
}

func (user User) TableName() string {
	return "sys_user"
}

func (user User) ToJwtUser() app.JwtUser {
	return app.JwtUser{
		Id: user.Id,
		// Ccid: user.Ccid,
		// Username:    user.Username,
		// Password:    user.Password,
		Application: setting.Config.App.Name,
	}
}

func (m User) Valid() User {
	// e.PanicIf(m.Username == "", "帐号不能为空！")
	// e.PanicIf(m.AccountType == 0, "帐号类型不能为空！")
	// e.PanicIf(len(m.Password) < 6 || m.Password == "", "密码6位以上，不能为空！")
	// e.PanicIf(m.Password != m.RePassword, "密码、确认密码不一致！")
	user := m.GetByAddr()
	e.PanicIf(user.Id > 0, "地址已注册！")
	return m
}

func (m User) ValidAccountPart() User {
	e.PanicIf(m.AccountPart&(1<<0) == 0, "请先完善基础信息：Email")
	e.PanicIf(m.AccountPart&(1<<1) == 0, "请先完善基础信息：头像")
	e.PanicIf(m.AccountPart&(1<<2) == 0, "请先完善基础信息：所在区域")
	return m
}

func (m User) SaveAccountPart() User {
	var part byte = 0
	if util.IsNotBlank(m.Email) {
		part |= 1 << 0 // 第一位 Email
	}
	if util.IsNotBlank(m.Avatar) {
		part |= 1 << 1 // 第二位 Avatar
	}
	if util.IsNotBlank(m.Country) && util.IsNotBlank(m.Province) && util.IsNotBlank(m.City) {
		part |= 1 << 2 // 第三位 区域
	}
	m.AccountPart = part

	return m
}

func (m User) Init() User {

	if m.Id == 0 {
		m.Id = uint64(util.IdWorker.Generate().Int64())
		// m.Password = util.MD5(m.Password)
	}
	return m
}

func (m User) Add() bool {
	create := domain.Db.Create(&m)
	return create.Error != nil
}

func (m *User) Update() error {
	m.SaveAccountPart() //保存完整度
	e := domain.Db.Model(&m).Updates(&m).Error
	if e != nil {
		log.Panicln("发生了错误", e.Error())
	}
	return e
}

func (entity User) GetPage(page *page2.UserPage) {
	var list []*User
	db := domain.Db
	if page.Nickname != "" {
		db = db.Where("name like ?", "%"+page.Nickname+"%")
	}
	//page查询封装
	page.SetModel(&User{}).
		SetRecords(&list).
		QueryPage(db)
}

func (user User) GetByAddr() *User {
	var u User
	/*r := */ domain.Db.Where("addr = ?", user.Addr).Find(&u)
	// if r.RowsAffected == 0 {
	// 	return nil
	// }
	return &u
}

func (user User) GetByAddrNotId() User {
	var u User
	domain.Db.Where("addr = ? and id<> ?", user.Addr, user.Id).Find(&u)
	return u
}

func (user User) QueryById() User {
	domain.Db.First(&user, user.Id)
	return user
}

func (m User) Delete() {
	domain.Db.Where("id = ?", m.Id).Delete(User{})
}

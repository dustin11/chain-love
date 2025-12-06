package sys

import (
	"chain-love/domain"
	"chain-love/domain/sys/enum"
	page2 "chain-love/domain/sys/page"
	"chain-love/pkg/app"
	"chain-love/pkg/e"
	"chain-love/pkg/util"
	"database/sql"
	"log"
)

type User struct {
	Id          uint64           `json:"id" gorm:"primary_key" form:"id"`
	Username    string           `json:"username" form:"username" gorm:"not null;size:52" example:"邮箱或手机号"` //form:后不能有空格，否则接不到值，坑爹
	AccountType enum.AccountType `json:"accountType" form:"accountType" gorm:"not null" `                   //binding:"required" 处理提示信息比较麻烦，所以不用
	Email       string           `json:"email" form:"email" example:"邮箱"`
	Mobile      string           `json:"mobile" form:"mobile" example:"手机号"`
	Nickname    string           `json:"nickname" form:"nickname" example:"昵称"`
	Password    string           `json:"password" form:"password" example:"密码"`
	RePassword  string           `json:"rePassword" form:"rePassword" gorm:"-"` //binding:"eqfield=Password"
	State       byte             `json:"state" example:"0"`
	Salt        sql.NullString   `json:"-" example:"盐值当头像"`
	Avatar      string           `json:"avatar" form:"avatar" example:"头像"`
	Ccid        int64            `json:"ccid" form:"ccid" example:"组织id"`
	domain.AreaModel

	RoleIds []int64 `json:"roleIds" form:"roleIds" gorm:"-"`
}

func (user User) TableName() string {
	return "sys_user"
}

func (user User) ToJwtUser() app.JwtUser {
	return app.JwtUser{
		Id:          user.Id,
		Ccid:        user.Ccid,
		Username:    user.Username,
		Password:    user.Password,
		Application: "SPIDER",
	}
}

func (m User) Valid() User {
	e.PanicIf(m.Username == "", "帐号不能为空！")
	e.PanicIf(m.AccountType == 0, "帐号类型不能为空！")
	e.PanicIf(len(m.Password) < 6 || m.Password == "", "密码6位以上，不能为空！")
	e.PanicIf(m.Password != m.RePassword, "密码、确认密码不一致！")
	user := m.GetByUsername()
	e.PanicIf(user.Id > 0, "用户名已被占用！")
	return m
}

func (m User) Init() User {
	// 通过帐号类型放入用户信息（使用 tagged switch 消除 staticcheck 警告）
	switch m.AccountType {
	case enum.Account_Type_Email:
		m.Email = m.Username
	case enum.Account_Type_Mobile:
		m.Mobile = m.Username
	}

	if m.Id == 0 {
		m.Id = uint64(util.IdWorker.Generate().Int64())
		m.Password = util.MD5(m.Password)
	}
	return m
}

func (m User) Add() bool {
	create := domain.Db.Create(&m)
	return create.Error != nil
}

func (m *User) Update() error {
	e := domain.Db.Model(&m).Updates(&m).Error
	if e != nil {
		log.Panicln("发生了错误", e.Error())
	}
	return e
}

func (entity User) GetPage(page *page2.UserPage) {
	var list []*User
	db := domain.Db
	if page.Username != "" {
		db = db.Where("name like ?", "%"+page.Username+"%")
	}
	//page查询封装
	page.SetModel(&User{}).
		SetRecords(&list).
		QueryPage(db)
}

func (user User) GetByUsername() User {
	domain.Db.Where("Username = ?", user.Username).Find(&user)
	return user
}

func (user User) GetByUsernameNotId() User {
	var u User
	domain.Db.Where("Username = ? and id<> ?", user.Username, user.Id).Find(&u)
	return u
}

func (user User) QueryById() User {
	domain.Db.First(&user, user.Id)
	return user
}

func (m User) Delete() {
	domain.Db.Where("id = ?", m.Id).Delete(User{})
}

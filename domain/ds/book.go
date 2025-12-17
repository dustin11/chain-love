package ds

import (
	"chain-love/domain"
	"chain-love/domain/ds/page"
	"chain-love/pkg/app/security"
	"log"
)

type Book struct {
	Id          int    `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Title       string `json:"title" form:"title" gorm:"type:varchar(255);not null;comment:标题"`
	Author      string `json:"author" form:"author" gorm:"type:varchar(100);comment:作者"`
	Translator  string `json:"translator" form:"translator" gorm:"type:varchar(100);comment:译者"`
	Language    string `json:"language" form:"language" gorm:"type:varchar(50);comment:语言"`
	CateId      int    `json:"cateId" form:"cateId" gorm:"type:int;comment:分类ID"`
	Addr        string `json:"addr" form:"addr" gorm:"type:varchar(64);comment:书籍合约地址"`
	Creator     string `json:"creator" form:"creator" gorm:"type:varchar(64);comment:创建者钱包地址"`
	Owner       string `json:"owner" form:"owner" gorm:"type:varchar(64);comment:拥有者钱包地址"`
	CoverImgUrl string `json:"coverImgUrl" form:"coverImgUrl" gorm:"type:varchar(255);comment:封面图片"`

	domain.Model
}

func (Book) TableName() string {
	return "ds_book"
}

func (m *Book) Init(user *security.JwtUser) *Book {
	m.Creator = user.Addr
	m.Owner = user.Addr
	return m
}

func (m *Book) Add() error {
	return domain.Db.Create(m).Error
}

func (m *Book) Update() error {
	err := domain.Db.Model(m).Updates(m).Error
	if err != nil {
		log.Println("Book Update Error", err)
	}
	return err
}

func (m Book) Delete() {
	domain.Db.Where("id = ?", m.Id).Delete(&Book{})
}

func (m Book) GetById() Book {
	var book Book
	domain.Db.First(&book, m.Id)
	return book
}

func (m Book) GetPage(page *page.BookPage) {
	var list []*Book
	db := domain.Db
	if page.Title != "" {
		db = db.Where("title like ?", "%"+page.Title+"%")
	}
	if page.Author != "" {
		db = db.Where("author like ?", "%"+page.Author+"%")
	}
	if page.Language != "" {
		db = db.Where("language = ?", page.Language)
	}
	if page.CateId > 0 {
		db = db.Where("cate_id = ?", page.CateId)
	}

	page.SetModel(&Book{}).
		SetRecords(&list).
		QueryPage(db)
}

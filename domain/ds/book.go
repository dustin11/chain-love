package ds

import (
	"errors"
	"log"

	"chain-love/domain"
	"chain-love/domain/ds/page"
	"chain-love/pkg/app/security"
)

type Book struct {
	Id         int    `json:"id" gorm:"primaryKey;autoIncrement" form:"id"`
	Name       string `json:"name" form:"name" gorm:"type:varchar(255);not null;comment:名称"`
	Author     string `json:"author" form:"author" gorm:"type:varchar(100);comment:作者"`
	Translator string `json:"trans" form:"trans" gorm:"type:varchar(100);comment:译者"`
	Language   string `json:"lang" form:"lang" gorm:"type:varchar(50);comment:语言"`
	CateId     int    `json:"cate,omitempty" form:"cate" gorm:"type:int;comment:分类ID"`
	Addr       string `json:"addr,omitempty" form:"addr" gorm:"type:varchar(64);comment:书籍合约地址"`
	Creator    string `json:"creator,omitempty" form:"creator" gorm:"type:varchar(64);comment:创建者钱包地址"`
	Owner      string `json:"owner,omitempty" form:"owner" gorm:"type:varchar(64);comment:拥有者钱包地址"`
	Cover      string `json:"cover,omitempty" form:"cover" gorm:"type:varchar(255);comment:封面图片"`

	FontStyle  FontStyle `json:"fontStyle,omitempty" gorm:"type:json;comment:字体样式"`
	AutoIndent bool      `json:"autoIndent,omitempty" form:"autoIndent" gorm:"comment:自动首行缩进"`
	Type       string    `json:"type" gorm:"type:varchar(50);comment:书籍类型"`
	PageCnt    int       `json:"pageCnt" form:"pageCnt" gorm:"type:int;comment:页数"`
	PlanetId   int       `json:"planetId" form:"planetId" gorm:"type:int;comment:星球ID"`
	Version    int       `json:"version" form:"version" gorm:"type:int;default:1;comment:版本号"`
	domain.CreatInfo
	// domain.UpdateInfo
}

func (Book) TableName() string {
	return "ds_book"
}

func (m *Book) Init(user *security.JwtUser) *Book {
	m.PlanetId = user.PlanetId
	m.Creator = user.Addr
	m.Owner = user.Addr
	m.CreatedBy = user.Id
	m.Version = 1
	return m
}

func (m *Book) Add() error {
	return domain.Db.Create(m).Error
}

// Update 修改：接收 userId 用于权限校验
func (m *Book) Update(userId uint64) error {
	// 增加 Where 条件：确保 ID 匹配且创建人是当前用户
	result := domain.Db.Where("id = ? AND created_by = ?", m.Id, userId).Updates(m)

	if result.Error != nil {
		log.Println("Book Update Error", result.Error)
		return result.Error
	}

	// 如果受影响行数为 0，说明记录不存在或无权修改（CreatedBy 不匹配）
	if result.RowsAffected == 0 {
		return errors.New("无权修改或记录不存在")
	}

	return nil
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
		db = db.Where("name like ?", "%"+page.Title+"%")
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

// GetIdsByPlanetId returns a list of book IDs associated with the given planet.
// If no books are found the slice will be empty and error will be nil.
func (m Book) GetIdByPlanetId(planetId int) ([]int, error) {
	var ids []int
	var books []Book
	res := domain.Db.Select("id").Where("planet_id = ?", planetId).Find(&books)
	if res.Error != nil {
		return nil, res.Error
	}
	for _, b := range books {
		ids = append(ids, b.Id)
	}
	return ids, nil
}

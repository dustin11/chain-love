package auth

import (
	"chain-love/domain"
	"chain-love/pkg/e"
	"time"

	"github.com/spruceid/siwe-go"
	"gorm.io/gorm"
)

const NonceDuration = 5 * time.Minute

// AuthNonce 存储生成的随机数，用于 SIWE 验证
type AuthNonce struct {
	Id        uint      `json:"id" gorm:"primary_key"`
	Address   string    `json:"address" gorm:"size:64"`
	Nonce     string    `json:"nonce" gorm:"size:64"`
	Used      bool      `json:"used"`
	ExpiresOn time.Time `json:"expires_on" gorm:"column:expires_on;type:timestamp(6)"`
	CreatedOn time.Time `json:"created_on" gorm:"column:created_on;type:timestamp(6)"`
}

func (AuthNonce) TableName() string {
	return "auth_nonce"
}

func (m AuthNonce) New(address string) *AuthNonce {
	m.Nonce = siwe.GenerateNonce()
	m.Address = address
	m.ExpiresOn = time.Now().Add(NonceDuration)
	m.CreatedOn = time.Now()
	return &m
}

func (n *AuthNonce) Add() *AuthNonce {
	er := domain.Db.Create(n).Error
	e.PanicIfServerErrTipMsg(er, "添加Nonce失败")
	return n
}

func GetValidNonce(address, nonce string) *AuthNonce {
	var rec AuthNonce
	err := domain.Db.Where("nonce = ? AND address = ? AND used = ? AND expires_on > ?", nonce, address, false, time.Now().UTC()).First(&rec).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			e.PanicIfParameterError(true, "Nonce无效或已过期")
		}
		e.PanicIfServerErrTipMsg(err, "查询Nonce失败")
	}
	return &rec
}

func (n *AuthNonce) MarkUsed() bool {
	err := domain.Db.Model(n).Update("used", true).Error
	e.PanicIfServerErrLogMsg(err, "更新Nonce状态失败")
	return true
}

func GenerateNonce(address string) *AuthNonce {
	return AuthNonce{}.New(address)
}

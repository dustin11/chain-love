package auth

import (
	"chain-love/domain"
	"chain-love/pkg/app/security"
	"chain-love/pkg/e"
	"chain-love/pkg/util"
	"time"

	"github.com/google/uuid"
)

const RefreshTokenDuration = 30 * 24 * time.Hour // 30天

// RefreshToken 存储刷新令牌，用于换取新的 Access Token
type RefreshToken struct {
	Id         uint      `json:"id" gorm:"primary_key"`
	TokenHash  string    `json:"token_hash" gorm:"size:128;index"`
	Address    string    `json:"address" gorm:"size:64"`
	ExpiresAt  time.Time `json:"expires_at"`
	Revoked    bool      `json:"revoked"` // 是否已撤销
	ClientIp   string    `json:"client_ip"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}

func (RefreshToken) TableName() string {
	return "auth_refresh_token"
}

func (m RefreshToken) New(address, clientIp, userAgent string) *RefreshToken {
	rawToken := uuid.NewString() + "." + security.MD5(util.RandomString(32)) // 简单生成随机串
	hash := security.SHA256(rawToken)

	m.TokenHash = hash
	m.Address = address
	m.ExpiresAt = time.Now().Add(RefreshTokenDuration)
	m.ClientIp = clientIp
	m.UserAgent = userAgent
	m.CreatedAt = time.Now()
	m.LastUsedAt = time.Now()

	return &m
}

func (m *RefreshToken) Add() *RefreshToken {
	er := domain.Db.Create(m).Error
	e.PanicIfServerErrLogMsg(er, "添加刷新令牌失败")
	return m
}

func FindValidRefreshByHash(hash string) (*RefreshToken, error) {
	var rec RefreshToken
	err := domain.Db.Where("token_hash = ? AND revoked = ? AND expires_at > ?", hash, false, time.Now().UTC()).First(&rec).Error
	return &rec, err
}

func RevokeRefreshByHash(hash string) error {
	return domain.Db.Model(&RefreshToken{}).Where("token_hash = ?", hash).Update("revoked", true).Error
}

func (r *RefreshToken) Revoke() error {
	return domain.Db.Model(r).Update("revoked", true).Error
}

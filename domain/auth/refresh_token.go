package auth

import (
	"senspace/domain"
	"senspace/pkg/app/security"
	"senspace/pkg/e"
	"senspace/pkg/util"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Refresh Token 默认有效期为 30 天。
const RefreshTokenDuration = 30 * 24 * time.Hour

// RefreshToken 存储刷新令牌，用于换取新的 Access Token
type RefreshToken struct {
	// Id 是数据库主键。
	Id uint `json:"id" gorm:"primary_key"`
	// TokenRaw 仅在生成时返回，不写入数据库。
	TokenRaw string `json:"-" gorm:"-"`
	// TokenHash 用于数据库检索和校验原始令牌。
	TokenHash string `json:"token_hash" gorm:"size:128;index"`
	// Address 标识令牌所属的钱包地址。
	Address string `json:"address" gorm:"size:64"`
	// ExpiresAt 是令牌的自然过期时间。
	ExpiresAt time.Time `json:"expires_at"`
	// Revoked 表示令牌已经登出或完成轮换。
	Revoked bool `json:"revoked"`
	// ClientIp 用于约束轮换宽限期内的并发请求。
	ClientIp string `json:"client_ip"`
	// UserAgent 用于约束轮换宽限期内的并发请求。
	UserAgent string `json:"user_agent"`
	// CreatedAt 是令牌创建时间。
	CreatedAt time.Time `json:"created_at"`
	// LastUsedAt 在撤销时记录轮换宽限窗口的起点。
	LastUsedAt time.Time `json:"last_used_at"`
}

func (RefreshToken) TableName() string {
	return "auth_refresh_token"
}

func (m RefreshToken) New(address, clientIp, userAgent string) *RefreshToken {
	rawToken := uuid.NewString() + "." + security.MD5(util.RandomString(32)) // 简单生成随机串
	hash := security.SHA256(rawToken)
	m.TokenRaw = rawToken
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
	er := m.AddWithDB(domain.Db)
	e.PanicIfServerErrLogMsg(er, "添加刷新令牌失败")
	return m
}

// 使用调用方提供的连接保存令牌，确保可加入外层事务。
func (m *RefreshToken) AddWithDB(db *gorm.DB) error {
	return db.Create(m).Error
}

func FindValidRefreshByHash(hash string) (*RefreshToken, error) {
	var rec RefreshToken
	err := domain.Db.Where("token_hash = ? AND revoked = ? AND expires_at > ?", hash, false, time.Now().UTC()).First(&rec).Error
	return &rec, err
}

// 在当前事务内锁定尚未自然过期的令牌，供轮换流程串行处理。
func FindRefreshByHashForUpdate(db *gorm.DB, hash string, now time.Time) (*RefreshToken, error) {
	var rec RefreshToken
	err := db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("token_hash = ? AND expires_at > ?", hash, now.UTC()).
		First(&rec).Error
	return &rec, err
}

func RevokeRefreshByHash(hash string) error {
	return domain.Db.Model(&RefreshToken{}).Where("token_hash = ?", hash).Update("revoked", true).Error
}

func (r *RefreshToken) Revoke() error {
	return r.RevokeWithDB(domain.Db, time.Now())
}

// 在当前事务内撤销令牌，并记录宽限窗口的起始时间。
func (r *RefreshToken) RevokeWithDB(db *gorm.DB, usedAt time.Time) error {
	return db.Model(r).Updates(map[string]interface{}{
		"revoked":      true,
		"last_used_at": usedAt,
	}).Error
}

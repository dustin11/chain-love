package auth

// // AuthNonce 存储生成的随机数，用于 SIWE 验证
// type AuthNonce struct {
// 	Id        uint64    `gorm:"primaryKey;autoIncrement"`
// 	Address   string    `gorm:"type:varchar(42);not null;index"` // 钱包地址
// 	Nonce     string    `gorm:"type:varchar(64);not null;unique"`
// 	ExpiresAt time.Time `gorm:"not null"`
// 	Used      bool      `gorm:"default:false"` // 是否已使用
// 	CreatedAt time.Time
// }

// // RefreshToken 存储刷新令牌，用于换取新的 Access Token
// type RefreshToken struct {
// 	Id          uint64    `gorm:"primaryKey;autoIncrement"`
// 	TokenHash   string    `gorm:"type:varchar(64);not null;unique;index"` // 存储 Token 的哈希值，不存明文
// 	Address     string    `gorm:"type:varchar(42);not null"`
// 	ExpiresAt   time.Time `gorm:"not null"`
// 	Revoked     bool      `gorm:"default:false"`    // 是否已撤销
// 	ParentToken string    `gorm:"type:varchar(64)"` // 用于追踪令牌轮换链（可选）
// 	CreatedAt   time.Time
// 	LastUsedAt  time.Time
// 	ClientIp    string `gorm:"type:varchar(45)"`
// 	UserAgent   string `gorm:"type:varchar(255)"`
// }

// 初始化表结构 (可在 main.go 或初始化逻辑中调用)
// func InitAuthTables() {
// 	domain.Db.AutoMigrate(&AuthNonce{}, &RefreshToken{})
// }

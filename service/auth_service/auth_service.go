package auth_service

import (
	"chain-love/domain/auth"
	"chain-love/domain/sys"
	"chain-love/pkg/app/security"
	"chain-love/pkg/e"

	"github.com/spruceid/siwe-go"
	"gorm.io/gorm"
)

// GenerateNonce 生成并保存 Nonce 到 MySQL
func GenerateNonce(address string) string {
	nonce := auth.GenerateNonce(address).Add()
	return nonce.Nonce
}

// VerifyAndLogin 验证 SIWE 签名并登录/注册
func VerifyAndLogin(messageStr, signature, clientIp, userAgent string) (string, string, *sys.User) {
	// 1. 解析消息
	msg, err := siwe.ParseMessage(messageStr)
	e.PanicIfParameterError(err != nil, "无效的消息格式")
	address := msg.GetAddress().String()
	nonce := msg.GetNonce()

	// 2. 验证 Nonce
	nonceRecord := auth.GetValidNonce(address, nonce)

	// 3. 验证签名
	_, err = msg.Verify(signature, nil, &nonce, nil)
	e.PanicIfServerErrTipMsg(err, "签名验证失败")

	// 4. 标记 Nonce 为已使用
	nonceRecord.MarkUsed()

	// 5. 获取用户
	user := sys.User{Addr: address}.GetByAddr()
	// 自动注册用户
	if user.Id == 0 {
		user := sys.User{}
		user.Init(address).Add()
	}

	// 6. 生成 Access Token (JWT)
	accessToken, err := security.GenerateToken(user.ToJwtUser())
	e.PanicIfServerErrLogMsg(err, "生成访问令牌失败")

	// 7. 生成 Refresh Token (Opaque Token) 并存入 MySQL
	refreshTokenRaw := createRefreshToken(address, clientIp, userAgent)

	return accessToken, refreshTokenRaw, user
}

// RefreshToken 刷新 Access Token (Token Rotation)
func RefreshToken(refreshTokenRaw, clientIp string) (string, string) {
	hash := security.SHA256(refreshTokenRaw)

	// 通过 model 查找
	record, err := auth.FindValidRefreshByHash(hash)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			e.PanicIfParameterError(true, "无效或过期的刷新令牌")
		}
		e.PanicIfServerErrLogMsg(err, "查询刷新令牌失败")
	}

	// 令牌轮换：撤销旧令牌
	if err := record.Revoke(); err != nil {
		e.PanicIfServerErrLogMsg(err, "撤销刷新令牌失败")
	}

	// 获取用户信息
	user := sys.User{Addr: record.Address}.GetByAddr()
	e.PanicIf(user.Id == 0, "用户不存在")

	// 颁发新 Access Token
	newAccess, err := security.GenerateToken(user.ToJwtUser())
	e.PanicIfServerErrLogMsg(err, "生成访问令牌失败")

	// 颁发新 Refresh Token
	newRefresh := createRefreshToken(record.Address, clientIp, record.UserAgent)

	return newAccess, newRefresh
}

// Logout 登出 (撤销 Refresh Token)
func Logout(refreshTokenRaw string) error {
	if refreshTokenRaw == "" {
		return e.ParameterError("缺少刷新令牌")
	}
	hash := security.SHA256(refreshTokenRaw)
	return auth.RevokeRefreshByHash(hash)
}

// 创建并存储 Refresh Token
func createRefreshToken(address, ip, ua string) string {
	m := auth.RefreshToken{}.New(address, ip, ua).Add()
	return m.TokenRaw
}

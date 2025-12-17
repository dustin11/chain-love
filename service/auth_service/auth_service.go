package auth_service

import (
	"chain-love/domain/auth"
	"chain-love/domain/sys"
	"chain-love/pkg/app/security"
	"chain-love/pkg/e"
	"chain-love/pkg/i18n"

	"github.com/spruceid/siwe-go"
	"gorm.io/gorm"
)

// GenerateNonce 生成并保存 Nonce 到 MySQL
func GenerateNonce(address string) string {
	nonce := auth.GenerateNonce(address).Add()
	return nonce.Nonce
}

// VerifyAndLogin 验证 SIWE 签名并登录/注册
func VerifyAndLogin(messageStr, signature, clientIp, userAgent, lang string) (string, string, *sys.User) {
	// 1. 解析消息
	msg, err := siwe.ParseMessage(messageStr)
	e.PanicIfParameterError(err != nil, i18n.Tr(lang, "auth.invalid_message_format"))
	address := msg.GetAddress().String()
	nonce := msg.GetNonce()

	// 2. 验证 Nonce
	nonceRecord := auth.GetValidNonce(address, nonce)

	// 3. 验证签名
	_, err = msg.Verify(signature, nil, &nonce, nil)
	e.PanicIfServerErrTipMsg(err, i18n.Tr(lang, "auth.signature_verification_failed"))

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
	e.PanicIfServerErrLogMsg(err, i18n.Tr(lang, "auth.generate_access_token_failed"))

	// 7. 生成 Refresh Token (Opaque Token) 并存入 MySQL
	refreshTokenRaw := createRefreshToken(address, clientIp, userAgent)

	return accessToken, refreshTokenRaw, user
}

// RefreshToken 刷新 Access Token (Token Rotation)
func RefreshToken(refreshTokenRaw, clientIp, lang string) (string, string) {
	hash := security.SHA256(refreshTokenRaw)

	// 通过 model 查找
	record, err := auth.FindValidRefreshByHash(hash)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			e.PanicIfParameterError(true, i18n.Tr(lang, "auth.invalid_or_expired_refresh_token"))
		}
		e.PanicIfServerErrLogMsg(err, i18n.Tr(lang, "auth.query_refresh_token_failed"))
	}

	// 令牌轮换：撤销旧令牌
	if err := record.Revoke(); err != nil {
		e.PanicIfServerErrLogMsg(err, i18n.Tr(lang, "auth.revoke_refresh_token_failed"))
	}

	// 获取用户信息
	user := sys.User{Addr: record.Address}.GetByAddr()
	e.PanicIf(user.Id == 0, i18n.Tr(lang, "auth.user_not_found"))

	// 颁发新 Access Token
	newAccess, err := security.GenerateToken(user.ToJwtUser())
	e.PanicIfServerErrLogMsg(err, i18n.Tr(lang, "auth.generate_access_token_failed"))

	// 颁发新 Refresh Token
	newRefresh := createRefreshToken(record.Address, clientIp, record.UserAgent)

	return newAccess, newRefresh
}

// Logout 登出 (撤销 Refresh Token)
func Logout(refreshTokenRaw, lang string) error {
	if refreshTokenRaw == "" {
		return e.ParameterError(i18n.Tr(lang, "auth.missing_refresh_token"))
	}
	hash := security.SHA256(refreshTokenRaw)
	return auth.RevokeRefreshByHash(hash)
}

// 创建并存储 Refresh Token
func createRefreshToken(address, ip, ua string) string {
	m := auth.RefreshToken{}.New(address, ip, ua).Add()
	return m.TokenRaw
}

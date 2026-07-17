package auth_service

import (
	"errors"
	"fmt"
	"senspace/domain"
	"senspace/domain/auth"
	"senspace/domain/sys"
	"senspace/pkg/app/security"
	"senspace/pkg/bizerr"
	"senspace/pkg/e"
	"senspace/pkg/i18n"
	"senspace/pkg/logging"
	"strings"
	"time"

	"github.com/spruceid/siwe-go"
	"gorm.io/gorm"
)

// 并发请求仅可在短窗口内复用刚撤销的令牌。
const refreshTokenRotationGrace = 10 * time.Second

var (
	// 事务回滚后统一映射为对外 401，避免泄露令牌或用户存在性。
	errInvalidRefreshToken = errors.New("invalid refresh token")
	errRefreshUserNotFound = errors.New("refresh token user not found")
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
	address := strings.ToLower(msg.GetAddress().String())
	nonce := msg.GetNonce()
	logging.Info("SIWE verify request address=", address, " domain=", msg.GetDomain(), " nonce=", nonce)

	// 2. 验证 Nonce
	nonceRecord := auth.GetValidNonce(address, nonce)

	// 3. 验证签名
	_, err = msg.Verify(signature, nil, &nonce, nil)
	if err != nil {
		logging.Error("SIWE signature verification failed address=", address, " domain=", msg.GetDomain(), " nonce=", nonce, " error=", err)
		e.PanicIfParameterError(true, i18n.Tr(lang, "auth.signature_verification_failed"))
	}

	// 4. 标记 Nonce 为已使用
	nonceRecord.MarkUsed()

	// 5. 获取用户
	user := sys.User{Addr: address}.GetByAddr()
	// 自动注册用户
	if user.Id == 0 {
		user = (&sys.User{}).Init(address)
		user.Add()
		logging.Info("SIWE auto registered user id=", user.Id, " address=", address)
	}

	// 6. 生成 Access Token (JWT)
	accessToken, err := security.GenerateToken(user.ToJwtUser())
	e.PanicIfServerErrLogMsg(err, i18n.Tr(lang, "auth.generate_access_token_failed"))

	// 7. 生成 Refresh Token (Opaque Token) 并存入 MySQL
	refreshTokenRaw := createRefreshToken(address, clientIp, userAgent)

	return accessToken, refreshTokenRaw, user
}

// RefreshToken 在事务内轮换访问令牌和刷新令牌。
func RefreshToken(refreshTokenRaw, clientIp, userAgent, lang string) (string, string, error) {
	hash := security.SHA256(refreshTokenRaw)
	var newAccess string
	var newRefresh string

	err := domain.Db.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		// 行锁保证同一 Refresh Token 的撤销和后继令牌创建按顺序执行。
		record, findErr := auth.FindRefreshByHashForUpdate(tx, hash, now)
		if findErr != nil {
			if errors.Is(findErr, gorm.ErrRecordNotFound) {
				return errInvalidRefreshToken
			}
			return fmt.Errorf("query refresh token: %w", findErr)
		}

		if record.Revoked {
			// 只接受同一客户端已经发出的并发请求，降低旧令牌重放风险。
			if !isRefreshTokenGraceReuseAllowed(record, clientIp, userAgent, now) {
				return errInvalidRefreshToken
			}
			logging.Info("refresh token grace reuse address=", record.Address)
		} else if revokeErr := record.RevokeWithDB(tx, now); revokeErr != nil {
			return fmt.Errorf("revoke refresh token: %w", revokeErr)
		}

		var user sys.User
		if userErr := tx.Where("addr = ?", record.Address).First(&user).Error; userErr != nil {
			if errors.Is(userErr, gorm.ErrRecordNotFound) {
				return errRefreshUserNotFound
			}
			return fmt.Errorf("query refresh user: %w", userErr)
		}

		access, generateErr := security.GenerateToken(user.ToJwtUser())
		if generateErr != nil {
			return fmt.Errorf("generate access token: %w", generateErr)
		}

		nextUserAgent := strings.TrimSpace(userAgent)
		if nextUserAgent == "" {
			nextUserAgent = record.UserAgent
		}
		next := auth.RefreshToken{}.New(record.Address, clientIp, nextUserAgent)
		// 新令牌与旧令牌撤销处于同一事务，任一步失败都会整体回滚。
		if addErr := next.AddWithDB(tx); addErr != nil {
			return fmt.Errorf("add refresh token: %w", addErr)
		}

		newAccess = access
		newRefresh = next.TokenRaw
		return nil
	})
	if err == nil {
		return newAccess, newRefresh, nil
	}
	if errors.Is(err, errInvalidRefreshToken) || errors.Is(err, errRefreshUserNotFound) {
		return "", "", bizerr.New(
			bizerr.KindUnauthorized,
			i18n.Tr(lang, "auth.invalid_or_expired_refresh_token"),
		)
	}
	return "", "", err
}

// 仅允许同 IP、同 User-Agent 在轮换宽限期内复用旧请求上下文。
func isRefreshTokenGraceReuseAllowed(record *auth.RefreshToken, clientIp, userAgent string, now time.Time) bool {
	if record == nil || !record.Revoked || record.LastUsedAt.IsZero() {
		return false
	}
	elapsed := now.Sub(record.LastUsedAt)
	if elapsed < 0 || elapsed > refreshTokenRotationGrace {
		return false
	}
	return strings.TrimSpace(record.ClientIp) == strings.TrimSpace(clientIp) &&
		strings.TrimSpace(record.UserAgent) == strings.TrimSpace(userAgent)
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

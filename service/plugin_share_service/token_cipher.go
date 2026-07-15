package plugin_share_service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"

	"senspace/pkg/setting"
)

// encryptManagementToken 只用于创建者管理页恢复分享链接，公开访问仍使用 TokenHash。
func encryptManagementToken(token string) (string, error) {
	block, err := newManagementCipher()
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, []byte(token), nil)
	return base64.RawURLEncoding.EncodeToString(sealed), nil
}

// decryptManagementToken 解密创建者管理页需要的分享令牌。
func decryptManagementToken(ciphertext string) (string, error) {
	block, err := newManagementCipher()
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	sealed, err := base64.RawURLEncoding.DecodeString(ciphertext)
	if err != nil || len(sealed) < gcm.NonceSize() {
		return "", errors.New("分享令牌密文无效")
	}
	nonce := sealed[:gcm.NonceSize()]
	plain, err := gcm.Open(nil, nonce, sealed[gcm.NonceSize():], nil)
	if err != nil {
		return "", errors.New("分享令牌密文无法解密")
	}
	return string(plain), nil
}

func newManagementCipher() (cipher.Block, error) {
	secret := setting.Config.App.JwtSecret
	if secret == "" {
		return nil, errors.New("分享令牌加密密钥未配置")
	}
	key := sha256.Sum256([]byte("senspace.plugin-share.management:" + secret))
	return aes.NewCipher(key[:])
}

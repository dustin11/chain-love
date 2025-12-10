package util

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
)

// IsBlank returns true if s is empty or contains only whitespace.
func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotBlank returns true if s contains non-whitespace characters.
func IsNotBlank(s string) bool {
	return !IsBlank(s)
}

// RandomString 生成一个长度为 n 的随机字节并返回其 hex 编码字符串（结果长度为 2*n）
// 调用方传入想要的字节数，例如 RandomString(32) 会返回 64 字符的 hex 字符串
func RandomString(n int) string {
	if n <= 0 {
		n = 32
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// 若需要统一错误处理请替换为项目内的 panic/日志方法
		panic(err)
	}
	return hex.EncodeToString(b)
}

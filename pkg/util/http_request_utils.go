package util

import (
	"errors"
	"net/http"
)

var (
	// ErrRequestBodyTooLarge 表示请求体超出允许大小。
	ErrRequestBodyTooLarge = errors.New("request body too large")
)

// LimitRequestBodyBytes 为请求体设置统一字节上限。
func LimitRequestBodyBytes(w http.ResponseWriter, r *http.Request, maxBytes int64) error {
	if w == nil || r == nil || maxBytes <= 0 {
		return nil
	}
	if r.ContentLength > maxBytes && r.ContentLength > 0 {
		return ErrRequestBodyTooLarge
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
	return nil
}

// IsRequestBodyTooLargeError 判断错误是否由请求体超限引起。
func IsRequestBodyTooLargeError(err error) bool {
	return err != nil && (errors.Is(err, ErrRequestBodyTooLarge) || err.Error() == "http: request body too large")
}

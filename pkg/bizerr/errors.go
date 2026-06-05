package bizerr

import (
	"errors"
	"net/http"
	"strings"

	"senspace/pkg/e"
)

// 用于区分参数、权限、缺失等业务错误。
type Kind string

const (
	// 请求参数不合法。
	KindParameter Kind = "parameter"
	// 当前请求缺少有效身份。
	KindUnauthorized Kind = "unauthorized"
	// 已登录但无权执行当前操作。
	KindForbidden Kind = "forbidden"
	// 目标数据不存在。
	KindNotFound Kind = "not_found"
	// 当前数据状态不允许本次操作。
	KindConflict Kind = "conflict"
)

// 统一给 service 返回、给 API 层识别并转换响应。
type Error struct {
	Kind    Kind
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

// 保留分类和文案，供上层统一处理。
func New(kind Kind, message string) error {
	return &Error{
		Kind:    kind,
		Message: strings.TrimSpace(message),
	}
}

// 便于 API 层按分类映射响应状态。
func IsKind(err error, kind Kind) bool {
	var bizErr *Error
	return errors.As(err, &bizErr) && bizErr.Kind == kind
}

// 未传文案时使用统一参数提示。
func Parameter(message string) error {
	if strings.TrimSpace(message) == "" {
		message = "参数错误"
	}
	return New(KindParameter, message)
}

func Unauthorized() error {
	return New(KindUnauthorized, "无授权信息！")
}

// 未传文案时使用统一权限提示。
func Forbidden(message string) error {
	if strings.TrimSpace(message) == "" {
		message = "无权限操作"
	}
	return New(KindForbidden, message)
}

// 未传文案时使用统一缺失提示。
func NotFound(message string) error {
	if strings.TrimSpace(message) == "" {
		message = "数据不存在"
	}
	return New(KindNotFound, message)
}

// 未传文案时使用统一冲突提示。
func Conflict(message string) error {
	if strings.TrimSpace(message) == "" {
		message = "数据状态冲突"
	}
	return New(KindConflict, message)
}

// 让 API 层直接复用统一错误到 HTTP 响应的转换。
func PanicHTTP(err error) {
	if err == nil {
		return
	}
	switch {
	case IsKind(err, KindUnauthorized):
		panic(e.UnauthorizedError(err.Error()))
	case IsKind(err, KindForbidden):
		panic(e.OtherError(err.Error()))
	case IsKind(err, KindNotFound):
		panic(e.NewError(http.StatusNotFound, 200404, err.Error()))
	case IsKind(err, KindParameter):
		panic(e.ParameterError(err.Error()))
	case IsKind(err, KindConflict):
		panic(e.NewError(http.StatusBadRequest, 200409, err.Error()))
	default:
		e.PanicServerErr(err)
	}
}

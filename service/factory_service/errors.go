package factory_service

// 服务层错误分类。
type ErrorKind string

const (
	// ErrorKindParameter 表示参数校验失败。
	ErrorKindParameter ErrorKind = "parameter"
	// ErrorKindForbidden 表示当前用户没有权限执行目标操作。
	ErrorKindForbidden ErrorKind = "forbidden"
	// ErrorKindNotFound 表示目标资源不存在。
	ErrorKindNotFound ErrorKind = "not_found"
	// ErrorKindConflict 表示当前业务状态与请求冲突。
	ErrorKindConflict ErrorKind = "conflict"
)

// 业务错误。
type ServiceError struct {
	Kind    ErrorKind
	Message string
}

// 错误消息。
func (e *ServiceError) Error() string {
	return e.Message
}

// 创建业务错误。
func newServiceError(kind ErrorKind, message string) error {
	return &ServiceError{Kind: kind, Message: message}
}

// 创建参数错误。
func newParameterError(message string) error {
	return newServiceError(ErrorKindParameter, message)
}

// 创建权限错误。
func newForbiddenError(message string) error {
	return newServiceError(ErrorKindForbidden, message)
}

// 创建不存在错误。
func newNotFoundError(message string) error {
	return newServiceError(ErrorKindNotFound, message)
}

// 创建冲突错误。
func newConflictError(message string) error {
	return newServiceError(ErrorKindConflict, message)
}

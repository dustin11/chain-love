package e

import (
	"chain-love/pkg/logging"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Error struct {
	StatusCode int         `json:"-"`
	Code       int         `json:"code"`
	Msg        string      `json:"msg"`
	Data       interface{} `json:"data" `
	//异常抛出位置，为排除通用抛出方法栈，直接定位到业务抛出点 通用异常处理时需要设置 如：error.go 非通用为空
	Locate string `json:"-"`
}

var (
	Success     = NewError(http.StatusOK, 0, "")
	ServerError = NewError(http.StatusInternalServerError, 200500, "系统异常，请稍后重试!")
	NotFound    = NewError(http.StatusNotFound, 200404, http.StatusText(http.StatusNotFound))
	AccessLimit = NewError(http.StatusUnauthorized, 200401, "please login first")
)

func ParameterError(message string) *Error {
	return NewError(http.StatusBadRequest, 200400, message)
}

func UnauthorizedError(message string) *Error {
	return NewError(http.StatusUnauthorized, 200401, message)
}

func serverError(message string) *Error {
	return NewError(http.StatusInternalServerError, 200500, message)
}

// 业务错误，跳过前端拦截器，方法内部自行处理
func BizError(message string, code int) *Error {
	return NewError(http.StatusOK, code, message)
}

func OtherError(message string) *Error {
	return NewError(http.StatusForbidden, 100403, message)
}

func otherErrorLocate(message string) *Error {
	return &Error{http.StatusForbidden, 100403, message, nil, logging.PROC_ERROR}
}

func SuccessData(data interface{}) *Error {
	return &Error{StatusCode: http.StatusOK, Code: 0, Data: data}
}

func (e *Error) Error() string {
	return e.Msg
}

func NewError(statusCode, Code int, msg string) *Error {
	return &Error{
		StatusCode: statusCode,
		Code:       Code,
		Msg:        msg,
	}
}

// 404处理
func HandleNotFound(c *gin.Context) {
	err := NotFound
	c.JSON(err.StatusCode, err)
}

func PanicIfUnauthorizedErr(b bool, msg string) {
	if b {
		e := UnauthorizedError(msg)
		panic(e)
	}
}

func PanicIfErr(err error) {
	if err != nil {
		e := OtherError(err.Error())
		panic(e)
	}
}

func PanicIfErrTipMsg(err error, tip string) {
	if err != nil {
		logging.Error(tip, err.Error())
		e := OtherError(tip)
		panic(e)
	}
}

func PanicIf(b bool, msg string) {
	if b {
		e := OtherError(msg)
		panic(e)
	}
}

func PanicMsg(msg string) {
	e := OtherError(msg)
	panic(e)
}

// 参数校验，失败则抛出 400 参数错误
func PanicIfParameterError(b bool, msg string) {
	if b {
		panic(ParameterError(msg))
	}
}

// 当发生内部/服务器错误时抛出 500 错误
func PanicIfServerErrLogMsg(err error, logMsg string) {
	if err != nil {
		logging.Error(logMsg + " - " + err.Error())
		panic(serverError("系统异常，请稍后重试!"))
	}
}

func PanicIfServerErrTipMsg(err error, msg string) {
	if err != nil {
		logging.Error(msg, err.Error())
		panic(serverError(msg))
	}
}

func PanicIfBizErr(b bool, msg string, code int) {
	if b {
		panic(BizError(msg, code))
	}
}

package app

type Result struct {
	Code    int         `json:"code" example:"1"`
	Message string      `json:"message" example:"响应信息"`
	Data    interface{} `json:"data" `
}

func OkResult() Result {
	return Result{1, "", nil}
}

func OkResultMsg(message string) Result {
	return Result{1, message, nil}
}

func OkResultMsgData(message string, data interface{}) Result {
	return Result{1, message, data}
}

func OkResultData(data interface{}) Result {
	return Result{1, "", data}
}

func FailResult() Result {
	return Result{-1, "", nil}
}
func FailResultMsg(message string) Result {
	return Result{-1, message, nil}
}

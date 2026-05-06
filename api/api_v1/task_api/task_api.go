package task_api

import (
	"strconv"

	"senspace/pkg/app"
	"senspace/pkg/app/contextx"
	"senspace/pkg/e"
	"senspace/service/task_service"
)

// 查询异步任务列表。
func ListTasks(ctx *contextx.AppContext) {
	result, err := task_service.ListTasks(task_service.Query{
		TaskType: ctx.Gin.Query("taskType"),
		Status:   ctx.Gin.Query("status"),
		Keyword:  ctx.Gin.Query("keyword"),
		Page:     parseQueryInt(ctx.Gin.Query("page"), 1),
		PageSize: parseQueryInt(ctx.Gin.Query("pageSize"), 20),
	})
	e.PanicServerErr(err)
	app.Response(ctx.Gin, e.SuccessData(result))
}

// 手动重试指定任务。
func RetryTask(ctx *contextx.AppContext) {
	result, err := task_service.RetryTaskNow(ctx.Gin.Param("id"))
	e.PanicServerErr(err)
	app.Response(ctx.Gin, e.SuccessData(result))
}

func parseQueryInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

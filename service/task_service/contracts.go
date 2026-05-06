package task_service

import "senspace/domain/task"

// Query 任务查询参数。
type Query struct {
	// 任务类型。
	TaskType string
	// 任务状态。
	Status string
	// 关键字，匹配 taskKey / bizType / lastError。
	Keyword string
	// 页码，从 1 开始。
	Page int
	// 每页数量。
	PageSize int
}

// RetryRequest 手动重试请求。
type RetryRequest struct {
	// 任务 ID。
	Id string `json:"id"`
}

// TaskView 前端展示视图。
type TaskView struct {
	Id             string      `json:"id"`
	TaskType       task.Type   `json:"taskType"`
	TaskKey        string      `json:"taskKey"`
	BizType        string      `json:"bizType,omitempty"`
	BizId          string      `json:"bizId,omitempty"`
	Status         task.Status `json:"status"`
	Priority       int         `json:"priority"`
	RetryCount     int         `json:"retryCount"`
	MaxRetry       int         `json:"maxRetry"`
	LastError      string      `json:"lastError,omitempty"`
	SourceVersion  string      `json:"sourceVersion,omitempty"`
	LeaseOwner     string      `json:"leaseOwner,omitempty"`
	StartedAt      string      `json:"startedAt,omitempty"`
	FinishedAt     string      `json:"finishedAt,omitempty"`
	NextRetryAt    string      `json:"nextRetryAt,omitempty"`
	CreatedAt      string      `json:"createdAt,omitempty"`
	UpdatedAt      string      `json:"updatedAt,omitempty"`
}

// TaskListResult 任务分页结果。
type TaskListResult struct {
	Total    int64      `json:"total"`
	Page     int        `json:"page"`
	PageSize int        `json:"pageSize"`
	Items    []TaskView `json:"items"`
}

package task

import (
	"time"

	"senspace/domain"
)

// AsyncTask 通用异步任务。
type AsyncTask struct {
	Id             int64  `json:"id,string" gorm:"primaryKey;autoIncrement:false;comment:任务ID"`
	TaskType       Type   `json:"taskType" gorm:"type:varchar(64);not null;index:idx_async_task_status_next,priority:3;index:idx_async_task_type_key,priority:1;comment:任务类型"`
	TaskKey        string `json:"taskKey" gorm:"type:varchar(255);not null;index:idx_async_task_type_key,priority:2;comment:任务目标键"`
	BizType        string `json:"bizType,omitempty" gorm:"type:varchar(64);not null;default:'';index:idx_async_task_biz,priority:1;comment:业务类型"`
	BizId          int64  `json:"bizId,string,omitempty" gorm:"not null;default:0;index:idx_async_task_biz,priority:2;comment:业务主键"`
	Status         Status `json:"status" gorm:"type:varchar(32);not null;index:idx_async_task_status_next,priority:1;comment:任务状态"`
	Priority       int    `json:"priority" gorm:"not null;default:100;index:idx_async_task_status_next,priority:4;comment:优先级，越小越高"`
	RetryCount     int    `json:"retryCount" gorm:"not null;default:0;comment:已重试次数"`
	MaxRetry       int    `json:"maxRetry" gorm:"not null;default:6;comment:最大重试次数"`
	LastError      string `json:"lastError,omitempty" gorm:"type:varchar(2000);not null;default:'';comment:最近一次错误"`
	PayloadJson    string `json:"payloadJson,omitempty" gorm:"type:json;comment:任务载荷"`
	ResultJson     string `json:"resultJson,omitempty" gorm:"type:json;comment:执行结果"`
	DedupeKey      string `json:"dedupeKey" gorm:"type:varchar(255);not null;uniqueIndex:uk_async_task_dedupe;comment:幂等去重键"`
	SourceVersion  string `json:"sourceVersion,omitempty" gorm:"type:varchar(128);not null;default:'';comment:任务依据版本"`
	LeaseOwner     string     `json:"leaseOwner,omitempty" gorm:"type:varchar(128);not null;default:'';comment:当前执行者"`
	LeaseExpiresAt *time.Time `json:"leaseExpiresAt,omitempty" gorm:"comment:执行租约过期时间"`
	NextRetryAt    *time.Time `json:"nextRetryAt,omitempty" gorm:"index:idx_async_task_status_next,priority:2;comment:下次可重试时间"`
	StartedAt      *time.Time `json:"startedAt,omitempty" gorm:"comment:开始执行时间"`
	FinishedAt     *time.Time `json:"finishedAt,omitempty" gorm:"comment:结束执行时间"`
	domain.CreatInfo
	domain.UpdateInfo
}

// TableName 表名。
func (AsyncTask) TableName() string {
	return "sys_async_task"
}

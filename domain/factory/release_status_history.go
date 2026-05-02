package factory

import (
	"time"

	"senspace/domain"
)

// 发布状态历史。
type ReleaseStatusHistory struct {
	Id             int64         `json:"id,string" gorm:"primaryKey;autoIncrement:false;comment:状态历史ID"`
	ReleaseId      int64         `json:"releaseId,string" gorm:"not null;index:idx_factory_status_history_release;comment:发布记录ID"`
	PluginId       string        `json:"-" gorm:"type:varchar(128);not null;index:idx_factory_status_history_plugin;comment:插件业务ID"`
	PreviousStatus ReleaseStatus `json:"previousStatus" gorm:"type:varchar(32);not null;comment:变更前状态"`
	NextStatus     ReleaseStatus `json:"nextStatus" gorm:"type:varchar(32);not null;comment:变更后状态"`
	Reason         string        `json:"reason,omitempty" gorm:"type:varchar(500);comment:状态变更原因"`
	ChangedBy      string        `json:"changedBy,omitempty" gorm:"type:varchar(64);comment:状态操作人"`
	ChangedAt      time.Time     `json:"changedAt" gorm:"autoCreateTime"`
	domain.CreatInfo
}

// TableName 表名。
func (ReleaseStatusHistory) TableName() string {
	return "fact_release_status_history"
}

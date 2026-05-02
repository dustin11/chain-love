package factory

import (
	"time"

	"senspace/domain"
)

// 发布价格历史。
type ReleasePriceHistory struct {
	Id                int64     `json:"id,string" gorm:"primaryKey;autoIncrement:false;comment:价格历史ID"`
	ReleaseId         int64     `json:"releaseId,string" gorm:"not null;index:idx_factory_price_history_release;comment:发布记录ID"`
	PluginId          string    `json:"-" gorm:"type:varchar(128);not null;index:idx_factory_price_history_plugin;comment:插件业务ID"`
	PreviousMintPrice string    `json:"previousMintPrice,omitempty" gorm:"type:decimal(36,18);comment:调价前价格"`
	NextMintPrice     string    `json:"nextMintPrice" gorm:"type:decimal(36,18);not null;comment:调价后价格"`
	Reason            string    `json:"reason,omitempty" gorm:"type:varchar(500);comment:调价原因"`
	ChangedBy         string    `json:"changedBy,omitempty" gorm:"type:varchar(64);comment:调价操作人"`
	ChangedAt         time.Time `json:"changedAt" gorm:"autoCreateTime"`
	domain.CreatInfo
}

// TableName 表名。
func (ReleasePriceHistory) TableName() string {
	return "fact_release_price_history"
}

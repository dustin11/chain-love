package factory

import (
	"time"

	"senspace/domain"
)

// 插件升级记录。
type UpgradeRecord struct {
	Id            int64       `json:"id,string" gorm:"primaryKey;autoIncrement:false;comment:升级记录ID"`
	OwnershipId   int64       `json:"ownershipId,string" gorm:"not null;index:idx_factory_upgrade_ownership;comment:资产记录ID"`
	UserId        uint64      `json:"userId,string" gorm:"type:bigint unsigned;not null;index:idx_factory_upgrade_user;comment:用户ID"`
	PluginId      string      `json:"pluginId" gorm:"type:varchar(128);not null;index:idx_factory_upgrade_plugin;comment:插件业务ID"`
	FromReleaseId int64       `json:"fromReleaseId,string" gorm:"not null;comment:升级前发布ID"`
	ToReleaseId   int64       `json:"toReleaseId,string" gorm:"not null;comment:升级后发布ID"`
	UpgradeType   UpgradeType `json:"upgradeType" gorm:"type:varchar(32);not null;comment:升级方式"`
	PaidAmount    string      `json:"paidAmount,omitempty" gorm:"type:decimal(36,18);comment:支付金额"`
	UpgradedAt    time.Time   `json:"upgradedAt" gorm:"autoCreateTime"`
	domain.CreatInfo
}

// TableName 表名。
func (UpgradeRecord) TableName() string {
	return "fact_upgrade_record"
}

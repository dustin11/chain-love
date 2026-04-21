package factory

import "time"

// 用户插件资产。
type UserOwnership struct {
	Id                 int64      `json:"id,string" gorm:"primaryKey;autoIncrement:false;comment:资产记录ID"`
	UserId             uint64     `json:"userId,string" gorm:"type:bigint unsigned;not null;uniqueIndex:idx_factory_ownership_user_plugin,priority:1;index:idx_factory_ownership_user;comment:用户ID"`
	PluginId           string     `json:"pluginId" gorm:"type:varchar(128);not null;uniqueIndex:idx_factory_ownership_user_plugin,priority:2;index:idx_factory_ownership_plugin;comment:插件业务ID"`
	MintedReleaseId    int64      `json:"mintedReleaseId,string" gorm:"not null;comment:首次铸造发布ID"`
	EffectiveReleaseId int64      `json:"effectiveReleaseId,string" gorm:"not null;comment:当前生效发布ID"`
	UpgradedAt         *time.Time `json:"upgradedAt,omitempty" gorm:"comment:最近升级时间"`
	CreatedAt          time.Time  `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt          time.Time  `json:"updatedAt" gorm:"autoUpdateTime"`
}

// TableName 表名。
func (UserOwnership) TableName() string {
	return "fact_user_ownership"
}

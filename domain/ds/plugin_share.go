package ds

import (
	"time"

	"senspace/domain"
)

// 插件分享状态。
type PluginShareStatus string

const (
	// 分享可正常访问。
	PluginShareStatusActive PluginShareStatus = "active"
	// 分享已由创建者撤销。
	PluginShareStatusRevoked PluginShareStatus = "revoked"
)

// 插件分享保存私有来源映射和公开快照，公开接口必须经过 service 投影。
type PluginShare struct {
	Id                     uint64               `json:"-" gorm:"primaryKey;autoIncrement;comment:分享ID"`
	TokenHash              string               `json:"-" gorm:"type:char(64);not null;uniqueIndex:uk_plugin_share_token_hash;comment:分享令牌哈希"`
	TokenCiphertext        string               `json:"-" gorm:"type:text;comment:创建者管理用分享令牌密文"`
	CreatorUserId          uint64               `json:"-" gorm:"not null;index:idx_plugin_share_creator;comment:创建用户ID"`
	SourcePlanetId         int                  `json:"-" gorm:"not null;index:idx_plugin_share_planet;comment:源星球ID"`
	SourcePluginInstanceId string               `json:"-" gorm:"type:varchar(128);not null;comment:源插件实例ID"`
	SourceSurfaceId        string               `json:"-" gorm:"type:varchar(128);comment:源活动表面ID"`
	ScopeKind              PluginAssetScopeKind `json:"-" gorm:"type:varchar(16);comment:资源空间类型"`
	ScopeOwnerKey          string               `json:"-" gorm:"type:varchar(128);comment:资源空间主人索引"`
	ScopeFactAssetId       *int64               `json:"-" gorm:"comment:源资产实例ID"`
	ScopePluginId          string               `json:"-" gorm:"type:varchar(128);comment:源插件业务ID"`
	ScopePluginVersion     string               `json:"-" gorm:"type:varchar(64);comment:源插件版本"`
	ScopeReleaseId         *int64               `json:"-" gorm:"comment:源发布记录ID"`
	ScopeDraftId           string               `json:"-" gorm:"type:varchar(128);comment:源草稿ID"`
	StateJson              string               `json:"-" gorm:"type:json;not null;comment:实例状态快照"`
	ResourceStateJson      string               `json:"-" gorm:"type:json;not null;comment:资源状态快照"`
	ResourceManifestJson   string               `json:"-" gorm:"type:json;not null;comment:公开资源清单投影"`
	ResourceMapJson        string               `json:"-" gorm:"type:json;not null;comment:资源别名私有映射"`
	PluginDescriptorJson   string               `json:"-" gorm:"type:json;not null;comment:插件加载描述"`
	CarrierStateJson       string               `json:"-" gorm:"type:json;not null;comment:承载链快照"`
	CameraStateJson        string               `json:"-" gorm:"type:json;not null;comment:相机快照"`
	MomentScope            string               `json:"-" gorm:"type:varchar(16);not null;default:plugin;index:idx_planet_moment_scope;comment:瞬间范围"`
	MomentText             string               `json:"-" gorm:"type:varchar(200);not null;default:'';comment:瞬间主题"`
	IsShared               bool                 `json:"-" gorm:"not null;index:idx_planet_moment_shared;comment:是否公开分享"`
	SnapshotKey            string               `json:"-" gorm:"type:varchar(128);comment:压缩快照文件键"`
	SnapshotHash           string               `json:"-" gorm:"type:char(64);comment:快照内容哈希"`
	QuotedMomentId         *uint64              `json:"-" gorm:"index:idx_planet_moment_quote;comment:引用瞬间ID"`
	QuotedMomentText       string               `json:"-" gorm:"type:varchar(200);comment:引用主题摘要"`
	Status                 PluginShareStatus    `json:"-" gorm:"type:varchar(32);not null;index:idx_plugin_share_status;comment:分享状态"`
	ExpiresAt              *time.Time           `json:"-" gorm:"index:idx_plugin_share_expires;comment:失效时间"`
	domain.CreatInfo
	domain.UpdateInfo
}

// TableName 表名。
func (PluginShare) TableName() string {
	return "ds_plugin_share"
}

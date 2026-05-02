package factory

import (
	"time"

	"senspace/domain"
)

// 发布记录。
type Release struct {
	Id               int64                  `json:"id,string" gorm:"primaryKey;autoIncrement:false;comment:发布记录ID"`
	PluginId         string                 `json:"pluginId" gorm:"type:varchar(128);not null;uniqueIndex:idx_factory_release_plugin_version,priority:1;index:idx_factory_release_plugin_current,priority:1;index:idx_factory_release_author,priority:2;comment:插件业务ID"`
	AuthorId         uint64                 `json:"authorId,omitempty,string" gorm:"type:bigint unsigned;not null;index:idx_factory_release_author,priority:1;comment:作者ID"`
	AuthorSnapshot   AuthorSnapshot         `json:"authorSnapshot,omitempty" gorm:"type:json;comment:作者快照"`
	Name             string                 `json:"name" gorm:"type:varchar(255);not null;comment:插件展示名称"`
	Version          string                 `json:"version" gorm:"type:varchar(64);not null;uniqueIndex:idx_factory_release_plugin_version,priority:2;comment:版本号"`
	Status           ReleaseStatus          `json:"status" gorm:"type:varchar(32);not null;index:idx_factory_release_status;comment:发布状态"`
	ReviewStatus     ReviewStatus           `json:"reviewStatus" gorm:"type:varchar(32);not null;default:'approved';comment:审核状态"`
	CurrentRelease   bool                   `json:"currentRelease" gorm:"not null;default:false;index:idx_factory_release_plugin_current,priority:2;comment:是否当前主推版本"`
	ManifestSnapshot PluginManifestSnapshot `json:"manifestSnapshot" gorm:"type:json;not null;comment:发布时锁定的manifest快照"`
	Summary          string                 `json:"summary" gorm:"type:varchar(1000);not null;comment:市场摘要"`
	Category         string                 `json:"category" gorm:"type:varchar(128);not null;index:idx_factory_release_category;comment:市场分类"`
	Tags             StringList             `json:"tags" gorm:"type:json;not null;comment:标签列表"`
	CoverUrl         string                 `json:"coverUrl,omitempty" gorm:"type:varchar(1024);comment:封面图地址"`
	TotalSupply      int64                  `json:"totalSupply" gorm:"not null;comment:总发行量"`
	MintPer          int64                  `json:"mintPer" gorm:"not null;comment:单次最大铸造量"`
	MintPrice        string                 `json:"mintPrice" gorm:"type:decimal(36,18);not null;comment:铸造价格"`
	MintedCount      int64                  `json:"mintedCount" gorm:"not null;default:0;comment:已铸造数量"`
	SourceHash       string                 `json:"sourceHash,omitempty" gorm:"type:varchar(128);comment:源码哈希"`
	BundleHash       string                 `json:"bundleHash,omitempty" gorm:"type:varchar(128);comment:构建包哈希"`
	Integrity        string                 `json:"integrity,omitempty" gorm:"type:varchar(255);comment:运行入口完整性校验值"`
	BuildStatus      BuildStatus            `json:"buildStatus,omitempty" gorm:"type:varchar(32);not null;default:'pending';comment:构建状态"`
	BuildError       string                 `json:"buildError,omitempty" gorm:"type:varchar(2000);comment:构建失败原因"`
	BuiltAt          *time.Time             `json:"builtAt,omitempty" gorm:"comment:最近一次构建完成时间"`
	RuntimeKind      ReleaseRuntimeKind     `json:"runtimeKind" gorm:"type:varchar(32);not null;default:'artifact';index:idx_factory_release_runtime;comment:运行来源类型"`
	UpgradePolicy    ReleaseUpgradePolicy   `json:"upgradePolicy,omitempty" gorm:"type:varchar(32);not null;default:'none';comment:升级策略"`
	UpgradePrice     string                 `json:"upgradePrice,omitempty" gorm:"type:decimal(36,18);comment:升级价格"`
	PublishedAt      *time.Time             `json:"publishedAt,omitempty" gorm:"comment:发布时间"`
	PausedAt         *time.Time             `json:"pausedAt,omitempty" gorm:"comment:暂停时间"`
	ClosedAt         *time.Time             `json:"closedAt,omitempty" gorm:"comment:关闭时间"`
	domain.CreatInfo
	domain.UpdateInfo
	// PriceHistory 价格历史。
	PriceHistory []ReleasePriceHistory `json:"priceHistory,omitempty" gorm:"-"`
}

// TableName 表名。
func (Release) TableName() string {
	return "fact_release"
}

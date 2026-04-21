package factory_service

import (
	"senspace/domain/factory"
	factoryvo "senspace/domain/factory/vo"
)

// 发布 manifest。
type PluginManifest struct {
	// 插件名称。
	Name string `json:"name"`
	// 插件版本号。
	Version string `json:"version"`
	// 插件入口文件。
	Entry string `json:"entry"`
	// 插件描述。
	Description string `json:"description,omitempty"`
}

// 可变市场信息。
type MutableMarketMetadata struct {
	// 市场摘要。
	Summary string `json:"summary"`
	// 市场分类。
	Category string `json:"category"`
	// 市场标签列表。
	Tags []string `json:"tags"`
	// 市场封面图地址。
	CoverUrl string `json:"coverUrl,omitempty"`
}

// 发布配置。
type ReleasePayload struct {
	MutableMarketMetadata
	// 当前版本总发行量。
	TotalSupply int64 `json:"totalSupply"`
	// 单次允许铸造的最大份数。
	MintPer int64 `json:"mintPer"`
	// 铸造价格，统一按字符串传输。
	MintPrice string `json:"mintPrice"`
	// 目标版本声明的升级策略。
	UpgradePolicy factory.ReleaseUpgradePolicy `json:"upgradePolicy,omitempty"`
	// 付费升级金额。
	UpgradePrice string `json:"upgradePrice,omitempty"`
}

// 发布请求。
type PublishRequest struct {
	// 插件业务 ID。
	PluginId string `json:"pluginId"`
	// 请求提交的 manifest 快照。
	Manifest PluginManifest `json:"manifest"`
	// 发布配置。
	Release ReleasePayload `json:"release"`
}

// 我的发布筛选。
type ReleaseQuery struct {
	PluginId    string
	Status      string
	CurrentOnly *bool
}

// 市场筛选。
type MarketQuery struct {
	PluginId    string
	Category    string
	Tags        []string
	Status      string
	CurrentOnly *bool
}

// 市场信息更新请求。
type UpdateReleaseRequest struct {
	Id     string                `json:"id"`
	Market MutableMarketMetadata `json:"market"`
}

// 价格更新请求。
type UpdateReleasePriceRequest struct {
	Id        string `json:"id"`
	MintPrice string `json:"mintPrice"`
	Reason    string `json:"reason,omitempty"`
}

// 状态更新请求。
type UpdateReleaseStatusRequest struct {
	Id     string                `json:"id"`
	Status factory.ReleaseStatus `json:"status"`
	Reason string                `json:"reason,omitempty"`
}

// 资产升级请求。
type UpgradeOwnershipRequest struct {
	Id          string `json:"id"`
	ToReleaseId string `json:"toReleaseId"`
}

// 铸造记录请求。
type RecordMintRequest struct {
	ReleaseId     string
	UserId        uint64
	WalletAddress string
	Quantity      int64
	TotalPaid     string
	ChainId       *int64
	TxHash        string
}

// 价格历史视图。
type PriceHistoryRecord struct {
	Id                string `json:"id"`
	ReleaseId         string `json:"releaseId"`
	PreviousMintPrice string `json:"previousMintPrice,omitempty"`
	NextMintPrice     string `json:"nextMintPrice"`
	Reason            string `json:"reason,omitempty"`
	ChangedBy         string `json:"changedBy,omitempty"`
	ChangedAt         string `json:"changedAt"`
}

// 发布记录视图。
type PublishRecord struct {
	Id             string                       `json:"id"`
	PluginId       string                       `json:"pluginId"`
	Name           string                       `json:"name"`
	Version        string                       `json:"version"`
	Status         factory.ReleaseStatus        `json:"status"`
	ReviewStatus   factory.ReviewStatus         `json:"reviewStatus,omitempty"`
	CurrentRelease bool                         `json:"currentRelease,omitempty"`
	TotalSupply    int64                        `json:"totalSupply"`
	MintedCount    int64                        `json:"mintedCount"`
	MintPrice      string                       `json:"mintPrice"`
	Summary        string                       `json:"summary,omitempty"`
	Category       string                       `json:"category,omitempty"`
	Tags           []string                     `json:"tags,omitempty"`
	CoverUrl       string                       `json:"coverUrl,omitempty"`
	SourceHash     string                       `json:"sourceHash,omitempty"`
	BundleHash     string                       `json:"bundleHash,omitempty"`
	UpgradePolicy  factory.ReleaseUpgradePolicy `json:"upgradePolicy,omitempty"`
	UpgradePrice   string                       `json:"upgradePrice,omitempty"`
	PublishedAt    string                       `json:"publishedAt,omitempty"`
	PausedAt       string                       `json:"pausedAt,omitempty"`
	ClosedAt       string                       `json:"closedAt,omitempty"`
	UpdatedAt      string                       `json:"updatedAt,omitempty"`
}

// 发布详情视图。
type PublishDetail struct {
	PublishRecord
	ManifestSnapshot PluginManifest       `json:"manifestSnapshot"`
	PriceHistory     []PriceHistoryRecord `json:"priceHistory,omitempty"`
}

// 用户插件资产视图。
type UserPluginOwnershipView = factoryvo.UserPluginOwnershipView

// 升级记录视图。
type UpgradeRecord struct {
	Id            string              `json:"id"`
	OwnershipId   string              `json:"ownershipId"`
	UserId        string              `json:"userId"`
	PluginId      string              `json:"pluginId"`
	FromReleaseId string              `json:"fromReleaseId"`
	ToReleaseId   string              `json:"toReleaseId"`
	UpgradeType   factory.UpgradeType `json:"upgradeType"`
	PaidAmount    string              `json:"paidAmount,omitempty"`
	UpgradedAt    string              `json:"upgradedAt"`
}

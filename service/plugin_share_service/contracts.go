package plugin_share_service

import (
	"encoding/json"
	"time"

	"senspace/domain/ds"
)

// SourceScopeInput 描述创建分享时使用的稳定资源空间，不接收 ownerKey。
type SourceScopeInput struct {
	Kind        ds.PluginAssetScopeKind `json:"kind"`
	FactAssetId int64                   `json:"factAssetId,string,omitempty"`
	PluginId    string                  `json:"pluginId,omitempty"`
	Version     string                  `json:"version,omitempty"`
	ReleaseId   int64                   `json:"releaseId,string,omitempty"`
	DraftId     string                  `json:"draftId,omitempty"`
}

// CreateInput 是创建单插件分享所需的宿主快照。
type CreateInput struct {
	SourcePlanetId   int               `json:"sourcePlanetId"`
	SourceInstanceId string            `json:"sourceInstanceId"`
	SourceSurfaceId  string            `json:"sourceSurfaceId,omitempty"`
	Scope            *SourceScopeInput `json:"scope,omitempty"`
	Plugin           json.RawMessage   `json:"plugin"`
	Carrier          json.RawMessage   `json:"carrier"`
	Camera           json.RawMessage   `json:"camera"`
	State            json.RawMessage   `json:"state,omitempty"`
	ResourceState    json.RawMessage   `json:"resourceState,omitempty"`
	ExpiresInHours   int               `json:"expiresInHours,omitempty"`
}

// CreateResult 返回独立分享入口。
type CreateResult struct {
	ShareToken string     `json:"shareToken"`
	ShareUrl   string     `json:"shareUrl"`
	ExpiresAt  *time.Time `json:"expiresAt,omitempty"`
}

// ShareListQuery 是创建者分享管理列表的查询参数。
type ShareListQuery struct {
	// 页码，从 1 开始。
	Page int
	// 每页数量，服务端限制最大值。
	PageSize int
	// 状态筛选，可为空；已删除记录不会返回。
	Status string
}

// ShareListItem 是只返回给创建者的分享管理视图。
type ShareListItem struct {
	// 分享记录 ID，不是公开令牌。
	Id string `json:"id"`
	// 插件显示名称，取自创建时的加载描述。
	PluginName string `json:"pluginName"`
	// 分享状态，只返回有效或已过期。
	Status string `json:"status"`
	// 可恢复时返回分享地址；旧记录可能为空。
	ShareUrl string `json:"shareUrl,omitempty"`
	// 创建时间。
	CreatedAt time.Time `json:"createdAt"`
	// 失效时间。
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// ShareListResult 是创建者分享管理列表结果。
type ShareListResult struct {
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"pageSize"`
	Items    []ShareListItem `json:"items"`
}

// Permissions 只公开当前访问者经过服务端计算后的权限结果。
type Permissions struct {
	IsPlanetOwner     bool `json:"isPlanetOwner"`
	CanManageComments bool `json:"canManageComments"`
	CanEditPlugin     bool `json:"canEditPlugin"`
}

// Bootstrap 是分享页唯一公开启动数据，不包含任何真实来源标识。
type Bootstrap struct {
	Schema           string          `json:"schema"`
	PluginInstanceId string          `json:"pluginInstanceId"`
	SurfaceId        string          `json:"surfaceId"`
	PlayerId         string          `json:"playerId,omitempty"`
	BackgroundUrl    string          `json:"backgroundUrl"`
	Plugin           json.RawMessage `json:"plugin"`
	Carrier          json.RawMessage `json:"carrier"`
	Camera           json.RawMessage `json:"camera"`
	State            json.RawMessage `json:"state"`
	ResourceState    json.RawMessage `json:"resourceState"`
	ResourceManifest json.RawMessage `json:"resourceManifest"`
	Permissions      Permissions     `json:"permissions"`
}

// ResourceTarget 是通过随机别名解析出的服务端文件。
type ResourceTarget struct {
	Path string
	Mime string
}

// ResourceManifest 是提供给只读分享资源管理器的公开资源列表。
type ResourceManifest struct {
	Schema string                 `json:"schema"`
	Assets []ResourceManifestItem `json:"assets"`
}

// ResourceManifestItem 使用分享作用域 ID 和代理 URL 描述资源。
type ResourceManifestItem struct {
	AssetId   string `json:"assetId"`
	Kind      string `json:"kind"`
	Mime      string `json:"mime"`
	Url       string `json:"url"`
	ThumbUrl  string `json:"thumbUrl,omitempty"`
	Hash      string `json:"hash,omitempty"`
	SizeBytes int64  `json:"sizeBytes,omitempty"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
}

// resourceMapEntry 仅在服务端保存，公开响应不可返回。
type resourceMapEntry struct {
	Path string `json:"path"`
	Mime string `json:"mime"`
}

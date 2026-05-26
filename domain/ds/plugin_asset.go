package ds

import "senspace/domain"

// 插件资源状态。
type PluginAssetStatus string

const (
	// 资源可正常使用。
	PluginAssetStatusActive PluginAssetStatus = "active"
	// 资源已删除，静态快照不再输出。
	PluginAssetStatusDeleted PluginAssetStatus = "deleted"
)

// 插件资源绑定状态。
type PluginAssetBindingStatus string

const (
	// 绑定可正常使用。
	PluginAssetBindingStatusActive PluginAssetBindingStatus = "active"
	// 绑定已删除，静态状态不再输出。
	PluginAssetBindingStatusDeleted PluginAssetBindingStatus = "deleted"
)

// 用户插件运行时上传的资源文件。
type PluginAsset struct {
	Id            uint64               `json:"id,string" gorm:"primaryKey;autoIncrement:false;comment:资源ID"`                                                                                            // 资源ID。
	ScopeKind     PluginAssetScopeKind `json:"scopeKind" gorm:"type:varchar(16);not null;index:idx_plugin_asset_scope_owner,priority:1;index:idx_plugin_asset_binding_scope,priority:1;comment:资源空间类型"` // 资源空间类型。
	OwnerKey      string               `json:"ownerKey" gorm:"type:varchar(128);not null;index:idx_plugin_asset_scope_owner,priority:2;comment:钱包索引键"`                                                  // 钱包索引键。
	OwnerAddress  string               `json:"ownerAddress,omitempty" gorm:"type:varchar(255);index:idx_plugin_asset_owner_addr;comment:钱包地址"`                                                          // 钱包地址。
	FactAssetId   *int64               `json:"factAssetId,string,omitempty" gorm:"index:idx_plugin_asset_fact;comment:插件资产实例ID"`                                                                        // 插件资产实例ID。
	PluginId      string               `json:"pluginId" gorm:"type:varchar(128);not null;index:idx_plugin_asset_dev,priority:1;comment:插件业务ID"`                                                         // 插件业务ID。
	PluginVersion string               `json:"pluginVersion" gorm:"type:varchar(64);not null;index:idx_plugin_asset_dev,priority:2;comment:插件版本号"`                                                      // 插件版本号。
	ReleaseId     *int64               `json:"releaseId,string,omitempty" gorm:"index:idx_plugin_asset_release;comment:发布记录ID"`                                                                         // 发布记录ID。
	Kind          string               `json:"kind" gorm:"type:varchar(32);not null;index:idx_plugin_asset_kind;comment:资源类型"`                                                                          // 资源类型。
	Mime          string               `json:"mime" gorm:"type:varchar(128);not null;comment:媒体类型"`                                                                                                     // 媒体类型。
	Hash          string               `json:"hash" gorm:"type:varchar(128);not null;index:idx_plugin_asset_hash;comment:内容哈希"`                                                                         // 内容哈希。
	SizeBytes     int64                `json:"sizeBytes" gorm:"not null;comment:文件大小"`                                                                                                                  // 文件大小。
	Width         int                  `json:"width,omitempty" gorm:"comment:图片宽度"`                                                                                                                     // 图片宽度。
	Height        int                  `json:"height,omitempty" gorm:"comment:图片高度"`                                                                                                                    // 图片高度。
	StoragePath   string               `json:"storagePath" gorm:"type:varchar(512);not null;comment:磁盘存储路径"`                                                                                            // 磁盘存储路径。
	PublicUrl     string               `json:"publicUrl" gorm:"type:varchar(512);not null;comment:公开访问地址"`                                                                                              // 公开访问地址。
	ThumbUrl      string               `json:"thumbUrl,omitempty" gorm:"type:varchar(512);comment:缩略图地址"`                                                                                               // 缩略图地址。
	Status        PluginAssetStatus    `json:"status" gorm:"type:varchar(32);not null;index:idx_plugin_asset_status;comment:资源状态"`                                                                      // 资源状态。
	domain.CreatInfo
	domain.UpdateInfo
}

// 数据表。
func (PluginAsset) TableName() string {
	return "ds_plugin_asset"
}

// 资源在插件实例中的展示配置。
type PluginAssetBinding struct {
	Id            uint64                   `json:"id,string" gorm:"primaryKey;autoIncrement;comment:绑定ID"`                                                       // 绑定ID。
	ScopeKind     PluginAssetScopeKind     `json:"scopeKind" gorm:"type:varchar(16);not null;index:idx_plugin_asset_binding_scope,priority:1;comment:资源空间类型"`    // 资源空间类型。
	OwnerKey      string                   `json:"ownerKey" gorm:"type:varchar(128);not null;index:idx_plugin_asset_binding_scope,priority:2;comment:钱包索引键"`     // 钱包索引键。
	FactAssetId   *int64                   `json:"factAssetId,string,omitempty" gorm:"index:idx_plugin_asset_binding_scope,priority:3;comment:插件资产实例ID"`         // 插件资产实例ID。
	PluginId      string                   `json:"pluginId" gorm:"type:varchar(128);not null;index:idx_plugin_asset_binding_scope,priority:4;comment:插件业务ID"`    // 插件业务ID。
	PluginVersion string                   `json:"pluginVersion" gorm:"type:varchar(64);not null;index:idx_plugin_asset_binding_scope,priority:5;comment:插件版本号"` // 插件版本号。
	AssetId       uint64                   `json:"assetId,string" gorm:"not null;index:idx_plugin_asset_binding_asset;comment:资源ID"`                             // 资源ID。
	CollectionKey string                   `json:"collectionKey" gorm:"type:varchar(128);not null;index:idx_plugin_asset_binding_collection;comment:资源集合键"`      // 资源集合键。
	SortOrder     int                      `json:"sortOrder" gorm:"not null;index:idx_plugin_asset_binding_order;comment:展示排序"`                                  // 展示排序。
	ConfigJson    string                   `json:"configJson,omitempty" gorm:"type:json;comment:展示配置JSON"`                                                       // 展示配置JSON。
	Status        PluginAssetBindingStatus `json:"status" gorm:"type:varchar(32);not null;index:idx_plugin_asset_binding_status;comment:绑定状态"`                   // 绑定状态。
	domain.CreatInfo
	domain.UpdateInfo
}

// 数据表。
func (PluginAssetBinding) TableName() string {
	return "ds_plugin_asset_binding"
}

// 插件实例整体运行状态。
type PluginInstanceState struct {
	Id            uint64               `json:"id,string" gorm:"primaryKey;autoIncrement;comment:状态ID"`                                                        // 状态ID。
	ScopeKind     PluginAssetScopeKind `json:"scopeKind" gorm:"type:varchar(16);not null;index:idx_plugin_instance_state_scope,priority:1;comment:资源空间类型"`    // 资源空间类型。
	OwnerKey      string               `json:"ownerKey" gorm:"type:varchar(128);not null;index:idx_plugin_instance_state_scope,priority:2;comment:钱包索引键"`     // 钱包索引键。
	FactAssetId   *int64               `json:"factAssetId,string,omitempty" gorm:"index:idx_plugin_instance_state_scope,priority:3;comment:插件资产实例ID"`         // 插件资产实例ID。
	PluginId      string               `json:"pluginId" gorm:"type:varchar(128);not null;index:idx_plugin_instance_state_scope,priority:4;comment:插件业务ID"`    // 插件业务ID。
	PluginVersion string               `json:"pluginVersion" gorm:"type:varchar(64);not null;index:idx_plugin_instance_state_scope,priority:5;comment:插件版本号"` // 插件版本号。
	SpaceId       string               `json:"spaceId,omitempty" gorm:"type:varchar(128);index:idx_plugin_instance_state_space;comment:空间ID"`                 // 空间ID。
	SurfaceId     string               `json:"surfaceId,omitempty" gorm:"type:varchar(128);comment:活动区域ID"`                                                   // 活动区域ID。
	PoseJson      string               `json:"poseJson,omitempty" gorm:"type:json;comment:位姿JSON"`                                                            // 位姿JSON。
	StateJson     string               `json:"stateJson,omitempty" gorm:"type:json;comment:实例状态JSON"`                                                         // 实例状态JSON。
	Revision      int64                `json:"revision" gorm:"not null;comment:状态版本号"`                                                                        // 状态版本号。
	domain.CreatInfo
	domain.UpdateInfo
}

// 数据表。
func (PluginInstanceState) TableName() string {
	return "ds_plugin_instance_state"
}

package factory

import "senspace/domain"

// 工厂 NFT 资产类型。
type AssetKind string

const (
	// 整件插件资产。
	AssetKindWhole AssetKind = "whole"
	// 组合资产中的组件。
	AssetKindComponent AssetKind = "component"
)

// 组件资产在组合结构中的角色。
type ComponentRole string

const (
	// 可作为组合挂载根节点。
	ComponentRoleRoot ComponentRole = "root"
	// 挂载到其它组件上的子节点。
	ComponentRoleChild ComponentRole = "child"
)

// 当前资产状态。
type AssetStatus string

const (
	// 有效。
	AssetStatusActive AssetStatus = "active"
	// 已转出。
	AssetStatusTransferred AssetStatus = "transferred"
	// 已销毁。
	AssetStatusBurned AssetStatus = "burned"
)

// NFT 权威资产状态。
type Asset struct {
	Id              int64              `json:"id,string" gorm:"primaryKey;autoIncrement:false;comment:资产ID"`
	PluginId        string             `json:"pluginId" gorm:"type:varchar(128);not null;index:idx_factory_asset_plugin;comment:插件业务ID"`
	ReleaseId       int64              `json:"releaseId,string" gorm:"not null;index:idx_factory_asset_release;comment:发布记录ID"`
	Version         string             `json:"version" gorm:"type:varchar(64);not null;comment:版本号"`
	RuntimeKind     ReleaseRuntimeKind `json:"runtimeKind" gorm:"type:varchar(32);not null;comment:运行来源类型"`
	AssetKind       AssetKind          `json:"assetKind" gorm:"type:varchar(64);not null;index:idx_factory_asset_kind;comment:资产类型"`
	CollectionKey   string             `json:"collectionKey,omitempty" gorm:"type:varchar(128);index:idx_factory_asset_collection;comment:集合业务键"`
	ComponentRole   ComponentRole      `json:"componentRole,omitempty" gorm:"type:varchar(32);index:idx_factory_asset_component_role;comment:组件角色"`
	ParentKey       string             `json:"parentKey,omitempty" gorm:"type:varchar(128);index:idx_factory_asset_parent_key;comment:父组件集合键"`
	TemplateRef     string             `json:"templateRef" gorm:"type:varchar(255);not null;comment:模板引用"`
	ItemId          string             `json:"itemId" gorm:"column:template_id;type:varchar(128);not null;comment:库存项ID"`
	ItemIndex       *int               `json:"itemIndex,omitempty" gorm:"index:idx_factory_asset_item_index;comment:模板项序号"`
	PluginOptions   string             `json:"pluginOptions,omitempty" gorm:"type:json;comment:插件属性面板参数"`
	Tier            string             `json:"tier,omitempty" gorm:"type:varchar(64);index:idx_factory_asset_tier;comment:资产等级"`
	TraitHash       string             `json:"traitHash,omitempty" gorm:"type:varchar(128);index:idx_factory_asset_trait;comment:冻结属性哈希"`
	OwnerAddress    string             `json:"ownerAddress" gorm:"type:varchar(255);not null;index:idx_factory_asset_owner;comment:当前钱包地址"`
	OwnerKey        string             `json:"ownerKey" gorm:"type:varchar(128);not null;index:idx_factory_asset_owner_key;comment:当前钱包索引键"`
	MintRecordId    int64              `json:"mintRecordId,string" gorm:"not null;index:idx_factory_asset_mint;comment:铸造记录ID"`
	ChainId         *int64             `json:"chainId,omitempty" gorm:"comment:链ID"`
	ContractAddress string             `json:"contractAddress,omitempty" gorm:"type:varchar(255);comment:合约地址"`
	TokenId         string             `json:"tokenId,omitempty" gorm:"type:varchar(255);comment:链上Token ID"`
	MetadataUri     string             `json:"metadataUri,omitempty" gorm:"type:varchar(512);comment:NFT metadata 静态地址"`
	ProofUri        string             `json:"proofUri,omitempty" gorm:"type:varchar(512);comment:NFT proof 静态地址"`
	Status          AssetStatus        `json:"status" gorm:"type:varchar(32);not null;index:idx_factory_asset_status;comment:资产状态"`
	domain.CreatInfo
	domain.UpdateInfo
}

// 数据表。
func (Asset) TableName() string {
	return "fact_asset"
}

// 组合关系状态。
type AssetRelationStatus string

const (
	// 生效。
	AssetRelationStatusActive AssetRelationStatus = "active"
	// 已移除。
	AssetRelationStatusRemoved AssetRelationStatus = "removed"
)

// NFT 之间的组合关系。
type AssetRelation struct {
	Id            int64               `json:"id,string" gorm:"primaryKey;autoIncrement:false;comment:关系ID"`
	OwnerKey      string              `json:"ownerKey" gorm:"type:varchar(128);not null;index:idx_factory_relation_owner;comment:钱包索引键"`
	RelationType  string              `json:"relationType" gorm:"type:varchar(64);not null;index:idx_factory_relation_type;comment:关系类型"`
	SourceAssetId int64               `json:"sourceAssetId,string" gorm:"not null;index:idx_factory_relation_source;comment:来源资产ID"`
	TargetAssetId int64               `json:"targetAssetId,string" gorm:"not null;index:idx_factory_relation_target;comment:目标资产ID"`
	MetadataJson  string              `json:"metadataJson,omitempty" gorm:"type:json;comment:关系扩展数据"`
	Status        AssetRelationStatus `json:"status" gorm:"type:varchar(32);not null;index:idx_factory_relation_status;comment:关系状态"`
	domain.CreatInfo
	domain.UpdateInfo
}

// 数据表。
func (AssetRelation) TableName() string {
	return "fact_asset_relation"
}

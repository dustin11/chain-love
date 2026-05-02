package factory

import (
	"time"

	"senspace/domain"

	"gorm.io/gorm"
)

// 库存池状态。
type NFTInventoryPoolStatus string

const (
	// 已冻结，可用于铸造。
	NFTInventoryPoolStatusFrozen NFTInventoryPoolStatus = "frozen"
	// 已激活，可用于公开铸造。
	NFTInventoryPoolStatusActive NFTInventoryPoolStatus = "active"
	// 库存已发完。
	NFTInventoryPoolStatusExhausted NFTInventoryPoolStatus = "exhausted"
	// 已关闭，不再发放。
	NFTInventoryPoolStatusClosed NFTInventoryPoolStatus = "closed"
)

// 库存发放策略。
type NFTInventoryStrategy string

const (
	// 按私有打乱顺序发放。
	NFTInventoryStrategyShuffled NFTInventoryStrategy = "shuffled"
	// 按元数据顺序发放。
	NFTInventoryStrategySequential NFTInventoryStrategy = "sequential"
	// 拍卖或人工分配。
	NFTInventoryStrategyAuction NFTInventoryStrategy = "auction"
	// 允许同一模板重复铸造。
	NFTInventoryStrategyAllowDuplicate NFTInventoryStrategy = "allowDuplicate"
)

// 单个库存项状态。
type NFTInventoryItemStatus string

const (
	// 可发放。
	NFTInventoryItemStatusAvailable NFTInventoryItemStatus = "available"
	// 已预留，等待支付或链上确认。
	NFTInventoryItemStatusReserved NFTInventoryItemStatus = "reserved"
	// 已发放。
	NFTInventoryItemStatusMinted NFTInventoryItemStatus = "minted"
	// 已销毁。
	NFTInventoryItemStatusBurned NFTInventoryItemStatus = "burned"
	// 已作废。
	NFTInventoryItemStatusVoided NFTInventoryItemStatus = "voided"
)

// 通用 NFT 库存池。
type NFTInventoryPool struct {
	Id             int64                  `json:"id,string" gorm:"primaryKey;autoIncrement:false;comment:库存池ID"`
	PluginId       string                 `json:"pluginId" gorm:"type:varchar(128);not null;index:idx_nft_inventory_pool_plugin;comment:插件业务ID"`
	ReleaseId      int64                  `json:"releaseId,string" gorm:"not null;uniqueIndex:idx_nft_inventory_pool_release_collection,priority:1;index:idx_nft_inventory_pool_release;comment:发布记录ID"`
	CollectionKey  string                 `json:"collectionKey" gorm:"type:varchar(128);not null;uniqueIndex:idx_nft_inventory_pool_release_collection,priority:2;comment:集合业务键"`
	AssetKind      AssetKind              `json:"assetKind" gorm:"type:varchar(64);not null;index:idx_nft_inventory_pool_kind;comment:资产类型"`
	MetadataRef    string                 `json:"metadataRef" gorm:"type:varchar(255);not null;comment:元数据来源"`
	Strategy       NFTInventoryStrategy   `json:"strategy" gorm:"type:varchar(32);not null;comment:发放策略"`
	TotalSupply    int64                  `json:"totalSupply" gorm:"not null;comment:库存总量"`
	MintedCount    int64                  `json:"mintedCount" gorm:"not null;default:0;comment:已发放数量"`
	Status         NFTInventoryPoolStatus `json:"status" gorm:"type:varchar(32);not null;index:idx_nft_inventory_pool_status;comment:库存池状态"`
	CollectionHash string                 `json:"collectionHash,omitempty" gorm:"type:varchar(128);comment:集合哈希"`
	MerkleRoot     string                 `json:"merkleRoot,omitempty" gorm:"type:varchar(128);comment:Merkle Root"`
	GeneratedAt    *time.Time             `json:"generatedAt,omitempty" gorm:"comment:生成时间"`
	FrozenAt       *time.Time             `json:"frozenAt,omitempty" gorm:"comment:冻结时间"`
	domain.CreatInfo
	domain.UpdateInfo
}

// 数据表。
func (NFTInventoryPool) TableName() string {
	return "fact_nft_inventory_pool"
}

// DropUnusedNFTInventoryPoolColumns 移除已废弃的库存池字段。
func DropUnusedNFTInventoryPoolColumns(db *gorm.DB) error {
	for _, column := range []string{"reserved_count", "seed_hash", "config_hash"} {
		if db.Migrator().HasColumn(&NFTInventoryPool{}, column) {
			if err := db.Migrator().DropColumn(&NFTInventoryPool{}, column); err != nil {
				return err
			}
		}
	}
	return nil
}

// 通用 NFT 库存项。
type NFTInventoryItem struct {
	Id            int64                  `json:"id,string" gorm:"primaryKey;autoIncrement:false;comment:库存项ID"`
	PoolId        int64                  `json:"poolId,string" gorm:"not null;uniqueIndex:idx_nft_inventory_item_pool_item,priority:1;index:idx_nft_inventory_item_pool_status_tier_shuffle,priority:1;comment:库存池ID"`
	PluginId      string                 `json:"pluginId" gorm:"type:varchar(128);not null;index:idx_nft_inventory_item_plugin;comment:插件业务ID"`
	ReleaseId     int64                  `json:"releaseId,string" gorm:"not null;uniqueIndex:idx_nft_inventory_item_release_collection_item,priority:1;index:idx_nft_inventory_item_release_collection_tier_status,priority:1;comment:发布记录ID"`
	CollectionKey string                 `json:"collectionKey" gorm:"type:varchar(128);not null;uniqueIndex:idx_nft_inventory_item_release_collection_item,priority:2;index:idx_nft_inventory_item_release_collection_tier_status,priority:2;comment:集合业务键"`
	AssetKind     AssetKind              `json:"assetKind" gorm:"type:varchar(64);not null;index:idx_nft_inventory_item_kind;comment:资产类型"`
	ItemId        string                 `json:"itemId" gorm:"type:varchar(128);not null;uniqueIndex:idx_nft_inventory_item_pool_item,priority:2;uniqueIndex:idx_nft_inventory_item_release_collection_item,priority:3;comment:元数据项ID"`
	ItemIndex     int64                  `json:"itemIndex" gorm:"not null;comment:元数据项序号"`
	Tier          string                 `json:"tier,omitempty" gorm:"type:varchar(64);index:idx_nft_inventory_item_pool_status_tier_shuffle,priority:3;index:idx_nft_inventory_item_release_collection_tier_status,priority:3;comment:等级"`
	TraitHash     string                 `json:"traitHash,omitempty" gorm:"type:varchar(128);index:idx_nft_inventory_item_trait;comment:属性哈希"`
	ShuffleIndex  int64                  `json:"shuffleIndex" gorm:"not null;index:idx_nft_inventory_item_pool_status_tier_shuffle,priority:4;comment:发放顺序"`
	MetadataHash  string                 `json:"metadataHash,omitempty" gorm:"type:varchar(128);comment:metadata 哈希"`
	LeafHash      string                 `json:"leafHash,omitempty" gorm:"type:varchar(128);comment:Merkle leaf 哈希"`
	ProofJson     string                 `json:"proofJson,omitempty" gorm:"type:json;comment:Merkle proof JSON"`
	Status        NFTInventoryItemStatus `json:"status" gorm:"type:varchar(32);not null;index:idx_nft_inventory_item_pool_status_tier_shuffle,priority:2;index:idx_nft_inventory_item_release_collection_tier_status,priority:4;comment:库存项状态"`
	ReservedBy    string                 `json:"reservedBy,omitempty" gorm:"type:varchar(128);comment:预留人"`
	ReservedAt    *time.Time             `json:"reservedAt,omitempty" gorm:"comment:预留时间"`
	ReservedUntil *time.Time             `json:"reservedUntil,omitempty" gorm:"comment:预留过期时间"`
	AssetId       *int64                 `json:"assetId,omitempty,string" gorm:"uniqueIndex:idx_nft_inventory_item_asset;comment:绑定资产ID"`
	TokenId       string                 `json:"tokenId,omitempty" gorm:"type:varchar(255);comment:链上Token ID"`
	MintRecordId  *int64                 `json:"mintRecordId,omitempty,string" gorm:"index:idx_nft_inventory_item_mint;comment:铸造记录ID"`
	OwnerKey      string                 `json:"ownerKey,omitempty" gorm:"type:varchar(128);index:idx_nft_inventory_item_owner;comment:持有人索引键"`
	MintedAt      *time.Time             `json:"mintedAt,omitempty" gorm:"comment:发放时间"`
	MetadataUri   string                 `json:"metadataUri,omitempty" gorm:"type:varchar(512);comment:NFT metadata 静态地址"`
	ProofUri      string                 `json:"proofUri,omitempty" gorm:"type:varchar(512);comment:NFT proof 静态地址"`
	domain.CreatInfo
	domain.UpdateInfo
}

// 数据表。
func (NFTInventoryItem) TableName() string {
	return "fact_nft_inventory_item"
}

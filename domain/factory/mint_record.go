package factory

import "time"

// NFT 铸造记录。
type MintRecord struct {
	Id            int64     `json:"id,string" gorm:"primaryKey;autoIncrement:false;comment:铸造记录ID"`
	PluginId      string    `json:"pluginId" gorm:"type:varchar(128);not null;index:idx_factory_mint_plugin;comment:插件业务ID"`
	ReleaseId     int64     `json:"releaseId,string" gorm:"not null;index:idx_factory_mint_release;comment:发布记录ID"`
	UserId        uint64    `json:"userId,string" gorm:"type:bigint unsigned;not null;index:idx_factory_mint_user;comment:用户ID"`
	WalletAddress string    `json:"walletAddress" gorm:"type:varchar(255);not null;comment:钱包地址"`
	Quantity      int64     `json:"quantity" gorm:"not null;comment:铸造数量"`
	TotalPaid     string    `json:"totalPaid" gorm:"type:decimal(36,18);not null;comment:支付总额"`
	ChainId       *int64    `json:"chainId,omitempty" gorm:"comment:链ID"`
	TxHash        string    `json:"txHash,omitempty" gorm:"type:varchar(255);comment:交易哈希"`
	MintedAt      time.Time `json:"mintedAt" gorm:"autoCreateTime"`
}

// TableName 表名。
func (MintRecord) TableName() string {
	return "fact_mint_record"
}

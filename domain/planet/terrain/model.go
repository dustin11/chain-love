package terrain

import "senspace/domain"

// Document 保存单个星球当前发布的地形文档。
type Document struct {
	// PlanetId 同时作为 planet 作用域和唯一主键。
	PlanetId int `json:"planetId" gorm:"column:planet_id;primaryKey;autoIncrement:false"`
	// SchemaVersion 标识状态 JSON 的结构版本。
	SchemaVersion int `json:"schemaVersion" gorm:"column:schema_version;not null"`
	// Revision 用于跨设备写入的乐观锁。
	Revision int64 `json:"revision" gorm:"column:revision;not null"`
	// StateJson 保存完整地形状态。
	StateJson string `json:"-" gorm:"column:state_json;type:longtext;not null"`
	// ContentHash 保存状态内容摘要。
	ContentHash string `json:"contentHash" gorm:"column:content_hash;type:char(64);not null"`
	domain.CreatInfo
	domain.UpdateInfo
}

// TableName 使用 planet 领域统一的 pla_ 前缀。
func (Document) TableName() string {
	return "pla_terrain_document"
}

// Tables 返回 planet/terrain 领域需要迁移的表。
func Tables() []interface{} {
	return []interface{}{&Document{}}
}

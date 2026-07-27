package terrain

import "encoding/json"

// 地形发布请求。
type SaveRequest struct {
	// ExpectedRevision 是客户端当前持有的服务端修订号。
	ExpectedRevision int64 `json:"expectedRevision"`
	// SchemaVersion 是地形状态结构版本。
	SchemaVersion int `json:"schemaVersion"`
	// State 是完整地形状态。
	State json.RawMessage `json:"state"`
}

// 前端可直接缓存的地形版本信封。
type DocumentResponse struct {
	// SchemaVersion 是状态结构版本。
	SchemaVersion int `json:"schemaVersion"`
	// Revision 是服务端修订号。
	Revision int64 `json:"revision"`
	// UpdatedAt 是最近发布时间。
	UpdatedAt string `json:"updatedAt"`
	// ContentHash 是内容摘要。
	ContentHash string `json:"contentHash"`
	// State 是完整地形状态。
	State json.RawMessage `json:"state"`
}

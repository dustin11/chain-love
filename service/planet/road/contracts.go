package road

import "encoding/json"

// SaveRequest 是道路发布请求。
type SaveRequest struct {
	// ExpectedRevision 是客户端当前持有的服务端修订号。
	ExpectedRevision int64 `json:"expectedRevision"`
	// SchemaVersion 是道路状态结构版本。
	SchemaVersion int `json:"schemaVersion"`
	// State 是完整道路网络状态。
	State json.RawMessage `json:"state"`
}

// DocumentResponse 是前端可直接缓存的道路版本信封。
type DocumentResponse struct {
	// SchemaVersion 是状态结构版本。
	SchemaVersion int `json:"schemaVersion"`
	// Revision 是服务端修订号。
	Revision int64 `json:"revision"`
	// UpdatedAt 是最近发布时间。
	UpdatedAt string `json:"updatedAt"`
	// ContentHash 是内容摘要。
	ContentHash string `json:"contentHash"`
	// State 是完整道路网络状态。
	State json.RawMessage `json:"state"`
}

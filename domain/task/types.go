package task

// 状态。
type Status string

const (
	// 待执行。
	StatusPending Status = "pending"
	// 执行中。
	StatusBuilding Status = "building"
	// 已完成。
	StatusReady Status = "ready"
	// 执行失败，等待重试。
	StatusFailed Status = "failed"
	// 超过最大重试次数，不再自动执行。
	StatusDead Status = "dead"
)

// 任务类型。
type Type string

const (
	// 发布任务。
	TypePublish Type = "publish"
	// 铸造任务。
	TypeMint Type = "mint"
)

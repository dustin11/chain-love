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
	// 工厂发布静态快照。
	TypeFactoryReleaseSnapshot Type = "factory_release_static_snapshot"
	// 工厂持有人资产快照。
	TypeFactoryOwnerAssetsSnapshot Type = "factory_owner_assets_snapshot"
)

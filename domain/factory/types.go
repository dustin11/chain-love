package factory

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// 发布状态。
type ReleaseStatus string

const (
	// 草稿。
	ReleaseStatusDraft ReleaseStatus = "draft"
	// 审核中。
	ReleaseStatusReviewing ReleaseStatus = "reviewing"
	// 已发布。
	ReleaseStatusPublished ReleaseStatus = "published"
	// 已暂停。
	ReleaseStatusPaused ReleaseStatus = "paused"
	// 已售罄。
	ReleaseStatusSoldOut ReleaseStatus = "sold_out"
	// 已关闭。
	ReleaseStatusClosed ReleaseStatus = "closed"
	// 已驳回。
	ReleaseStatusRejected ReleaseStatus = "rejected"
	// 已归档。
	ReleaseStatusArchived ReleaseStatus = "archived"
)

// 审核状态。
type ReviewStatus string

const (
	// 待审核。
	ReviewStatusPending ReviewStatus = "pending"
	// 已通过。
	ReviewStatusApproved ReviewStatus = "approved"
	// 已拒绝。
	ReviewStatusRejected ReviewStatus = "rejected"
)

// 构建状态。
type BuildStatus string

const (
	// 待构建。
	BuildStatusPending BuildStatus = "pending"
	// 构建中。
	BuildStatusBuilding BuildStatus = "building"
	// 构建成功。
	BuildStatusReady BuildStatus = "ready"
	// 构建失败。
	BuildStatusFailed BuildStatus = "failed"
)

// 升级策略。
type ReleaseUpgradePolicy string

const (
	// 不可升级。
	ReleaseUpgradePolicyNone ReleaseUpgradePolicy = "none"
	// 免费升级。
	ReleaseUpgradePolicyFree ReleaseUpgradePolicy = "free"
	// 付费升级。
	ReleaseUpgradePolicyPaid ReleaseUpgradePolicy = "paid"
	// 大版本付费。
	ReleaseUpgradePolicyMajorPaid ReleaseUpgradePolicy = "major_paid"
)

// 升级方式。
type UpgradeType string

const (
	// 免费。
	UpgradeTypeFree UpgradeType = "free"
	// 付费。
	UpgradeTypePaid UpgradeType = "paid"
	// 平台赠送。
	UpgradeTypeAdminGrant UpgradeType = "admin_grant"
)

// 资产升级状态。
type OwnershipUpgradeState string

const (
	// 已最新。
	OwnershipUpgradeStateUpToDate OwnershipUpgradeState = "up_to_date"
	// 可升级。
	OwnershipUpgradeStateUpgradable OwnershipUpgradeState = "upgradable"
	// 需付费升级。
	OwnershipUpgradeStateUpgradeRequired OwnershipUpgradeState = "upgrade_required"
)

// manifest 快照。
type PluginManifestSnapshot struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Entry       string `json:"entry"`
	Description string `json:"description,omitempty"`
}

// 作者快照。
type AuthorSnapshot struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar,omitempty"`
}

// 转 JSON。
func (m PluginManifestSnapshot) Value() (driver.Value, error) {
	if isEmptyJSONStruct(m) {
		return "{}", nil
	}
	bytes, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

// 读数据库值。
func (m *PluginManifestSnapshot) Scan(value interface{}) error {
	return scanJSONValue(value, m)
}

// 转 JSON。
func (a AuthorSnapshot) Value() (driver.Value, error) {
	if isEmptyJSONStruct(a) {
		return "{}", nil
	}
	bytes, err := json.Marshal(a)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

// 读数据库值。
func (a *AuthorSnapshot) Scan(value interface{}) error {
	return scanJSONValue(value, a)
}

// 字符串列表。
type StringList []string

// 转 JSON 数组。
func (s StringList) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	bytes, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return string(bytes), nil
}

// 读数据库值。
func (s *StringList) Scan(value interface{}) error {
	if s == nil {
		return fmt.Errorf("factory.StringList: nil receiver")
	}
	if value == nil {
		*s = StringList{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		if len(v) == 0 {
			*s = StringList{}
			return nil
		}
		return json.Unmarshal(v, s)
	case string:
		if v == "" {
			*s = StringList{}
			return nil
		}
		return json.Unmarshal([]byte(v), s)
	default:
		return fmt.Errorf("factory.StringList: unsupported scan type %T", value)
	}
}

// 兼容读取 JSON。
func scanJSONValue(value interface{}, target interface{}) error {
	switch v := value.(type) {
	case nil:
		return nil
	case []byte:
		if len(v) == 0 {
			return nil
		}
		return json.Unmarshal(v, target)
	case string:
		if v == "" {
			return nil
		}
		return json.Unmarshal([]byte(v), target)
	default:
		return fmt.Errorf("factory: unsupported json scan type %T", value)
	}
}

// 判断空对象。
func isEmptyJSONStruct(v interface{}) bool {
	bytes, err := json.Marshal(v)
	if err != nil {
		return false
	}
	return string(bytes) == "{}"
}

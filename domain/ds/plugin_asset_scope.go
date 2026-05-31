package ds

import (
	"errors"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var pluginAssetSafeSegmentPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]{0,127}$`)

// 插件资源空间类型。
type PluginAssetScopeKind string

const (
	// 已发布并已拥有的工厂资产实例空间。
	PluginAssetScopeFact PluginAssetScopeKind = "fact"
	// 开发工作区插件实例空间。
	PluginAssetScopeDev PluginAssetScopeKind = "dev"
	// 铸造前配置草稿空间。
	PluginAssetScopeDraft PluginAssetScopeKind = "draft"
)

// 插件资源空间。
type PluginAssetScope struct {
	Kind          PluginAssetScopeKind `json:"kind"`                         // 资源空间类型。
	OwnerKey      string               `json:"ownerKey"`                     // 钱包索引键。
	OwnerAddress  string               `json:"ownerAddress,omitempty"`       // 钱包地址。
	FactAssetId   int64                `json:"factAssetId,string,omitempty"` // 工厂资产实例ID。
	PluginId      string               `json:"pluginId,omitempty"`           // 插件业务ID。
	PluginVersion string               `json:"version,omitempty"`            // 插件版本号。
	ReleaseId     int64                `json:"releaseId,string,omitempty"`   // 发布记录ID。
	DraftId       string               `json:"draftId,omitempty"`            // 草稿ID。
}

// Validate 校验资源空间是否合法。
func (s PluginAssetScope) Validate() error {
	ownerKey := strings.TrimSpace(s.OwnerKey)
	if ownerKey == "" {
		return errors.New("ownerKey不能为空")
	}
	if err := ValidatePluginAssetPathSegment("ownerKey", ownerKey); err != nil {
		return err
	}
	switch s.Kind {
	case PluginAssetScopeFact:
		if s.FactAssetId <= 0 {
			return errors.New("factAssetId不能为空")
		}
	case PluginAssetScopeDev:
		if err := ValidatePluginAssetPathSegment("pluginId", s.PluginId); err != nil {
			return err
		}
		if err := ValidatePluginAssetPathSegment("pluginVersion", s.PluginVersion); err != nil {
			return err
		}
	case PluginAssetScopeDraft:
		if s.ReleaseId <= 0 {
			return errors.New("releaseId不能为空")
		}
		if err := ValidatePluginAssetPathSegment("draftId", s.DraftId); err != nil {
			return err
		}
	default:
		return errors.New("scope kind无效")
	}
	return nil
}

// ValidatePluginAssetPathSegment 校验插件资源路径中的单个目录片段。
func ValidatePluginAssetPathSegment(label string, value string) error {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return errors.New(label + "不能为空")
	}
	if trimmed == "." || trimmed == ".." {
		return errors.New(label + "无效")
	}
	if filepath.Clean(trimmed) != trimmed {
		return errors.New(label + "不能包含路径分隔符")
	}
	if strings.ContainsAny(trimmed, `/\`) {
		return errors.New(label + "不能包含路径分隔符")
	}
	if !pluginAssetSafeSegmentPattern.MatchString(trimmed) {
		return errors.New(label + "包含非法字符")
	}
	return nil
}

// StaticPathParts 返回 scope 对应的静态目录层级。
func (s PluginAssetScope) StaticPathParts() []string {
	switch s.Kind {
	case PluginAssetScopeFact:
		return []string{
			string(PluginAssetScopeFact),
			strconv.FormatInt(s.FactAssetId, 10),
		}
	case PluginAssetScopeDev:
		return []string{
			string(PluginAssetScopeDev),
			strings.TrimSpace(s.OwnerKey),
			strings.TrimSpace(s.PluginId),
			strings.TrimSpace(s.PluginVersion),
		}
	case PluginAssetScopeDraft:
		return []string{
			string(PluginAssetScopeDraft),
			strings.TrimSpace(s.OwnerKey),
			strconv.FormatInt(s.ReleaseId, 10),
			strings.TrimSpace(s.DraftId),
		}
	default:
		return []string{}
	}
}

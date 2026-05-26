package ds

import (
	"errors"
	"path/filepath"
	"strconv"
	"strings"
)

// 插件资源空间类型。
type PluginAssetScopeKind string

const (
	// 已发布并已拥有的工厂资产实例空间。
	PluginAssetScopeFact PluginAssetScopeKind = "fact"
	// 开发工作区插件实例空间。
	PluginAssetScopeDev PluginAssetScopeKind = "dev"
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
}

// Validate 校验资源空间是否合法。
func (s PluginAssetScope) Validate() error {
	ownerKey := strings.TrimSpace(s.OwnerKey)
	if ownerKey == "" {
		return errors.New("ownerKey不能为空")
	}
	switch s.Kind {
	case PluginAssetScopeFact:
		if s.FactAssetId <= 0 {
			return errors.New("factAssetId不能为空")
		}
	case PluginAssetScopeDev:
		if strings.TrimSpace(s.PluginId) == "" {
			return errors.New("pluginId不能为空")
		}
		if strings.TrimSpace(s.PluginVersion) == "" {
			return errors.New("pluginVersion不能为空")
		}
	default:
		return errors.New("scope kind无效")
	}
	return nil
}

// StaticPathParts 返回 scope 对应的静态目录层级。
func (s PluginAssetScope) StaticPathParts() []string {
	switch s.Kind {
	case PluginAssetScopeFact:
		return []string{
			string(PluginAssetScopeFact),
			strings.TrimSpace(s.OwnerKey),
			strconv.FormatInt(s.FactAssetId, 10),
		}
	case PluginAssetScopeDev:
		return []string{
			string(PluginAssetScopeDev),
			strings.TrimSpace(s.OwnerKey),
			filepath.Clean(strings.TrimSpace(s.PluginId)),
			filepath.Clean(strings.TrimSpace(s.PluginVersion)),
		}
	default:
		return []string{}
	}
}

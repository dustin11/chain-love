package ds

import (
	"path/filepath"
	"strconv"
	"strings"

	"senspace/pkg/setting"
)

// 插件资源静态根目录。
func PluginAssetsRoot() string {
	root := strings.TrimSpace(setting.Config.App.FilePath.PluginAssets)
	if root != "" {
		return root
	}
	if base := strings.TrimSpace(setting.Config.App.RuntimeRootPath); base != "" {
		return filepath.Join(base, "plugin-assets")
	}
	return filepath.Join("runtime", "plugin-assets")
}

// 插件实例资源目录。
func PluginAssetInstanceDir(ownerKey string, factAssetId int64) string {
	return filepath.Join(PluginAssetsRoot(), ownerKey, strconv.FormatInt(factAssetId, 10))
}

// 单个资源文件目录。
func PluginAssetDir(ownerKey string, factAssetId int64, assetId uint64) string {
	return filepath.Join(
		PluginAssetInstanceDir(ownerKey, factAssetId),
		"assets",
		strconv.FormatUint(assetId, 10),
	)
}

// 资源清单文件路径。
func PluginAssetManifestPath(ownerKey string, factAssetId int64) string {
	return filepath.Join(PluginAssetInstanceDir(ownerKey, factAssetId), "manifest.json")
}

// 插件状态文件路径。
func PluginAssetStatePath(ownerKey string, factAssetId int64) string {
	return filepath.Join(PluginAssetInstanceDir(ownerKey, factAssetId), "state.json")
}

// 插件资源静态 URL。
func PluginAssetStaticURL(pathParts ...string) string {
	parts := append([]string{"static", "plugin-assets"}, pathParts...)
	return "/" + filepath.ToSlash(filepath.Join(parts...))
}

// 资源清单静态 URL。
func PluginAssetManifestURL(ownerKey string, factAssetId int64) string {
	return PluginAssetStaticURL(ownerKey, strconv.FormatInt(factAssetId, 10), "manifest.json")
}

// 插件状态静态 URL。
func PluginAssetStateURL(ownerKey string, factAssetId int64) string {
	return PluginAssetStaticURL(ownerKey, strconv.FormatInt(factAssetId, 10), "state.json")
}

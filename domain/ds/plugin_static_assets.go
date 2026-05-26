package ds

import (
	"errors"
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
func PluginAssetInstanceDir(scope PluginAssetScope) string {
	return filepath.Join(append([]string{PluginAssetsRoot()}, scope.StaticPathParts()...)...)
}

// 单个资源文件目录。
func PluginAssetDir(scope PluginAssetScope, assetId uint64) string {
	return filepath.Join(
		PluginAssetInstanceDir(scope),
		"assets",
		strconv.FormatUint(assetId, 10),
	)
}

// 资源清单文件路径。
func PluginAssetManifestPath(scope PluginAssetScope) string {
	return filepath.Join(PluginAssetInstanceDir(scope), "manifest.json")
}

// 插件状态文件路径。
func PluginAssetStatePath(scope PluginAssetScope) string {
	return filepath.Join(PluginAssetInstanceDir(scope), "state.json")
}

// 插件资源静态 URL。
func PluginAssetStaticURL(pathParts ...string) string {
	parts := append([]string{"static", "plugin-assets"}, pathParts...)
	return "/" + filepath.ToSlash(filepath.Join(parts...))
}

// 资源清单静态 URL。
func PluginAssetManifestURL(scope PluginAssetScope) string {
	return PluginAssetStaticURL(append(scope.StaticPathParts(), "manifest.json")...)
}

// 插件状态静态 URL。
func PluginAssetStateURL(scope PluginAssetScope) string {
	return PluginAssetStaticURL(append(scope.StaticPathParts(), "state.json")...)
}

// MustValidatePluginAssetScope 确保 scope 在生成静态路径前已合法。
func MustValidatePluginAssetScope(scope PluginAssetScope) {
	if err := scope.Validate(); err != nil {
		panic(errors.New("invalid plugin asset scope: " + err.Error()))
	}
}

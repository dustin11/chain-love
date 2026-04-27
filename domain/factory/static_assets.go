package factory

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"senspace/pkg/setting"
)

const factoryOwnerHashSalt = "senspace-owner-v1:"

// 发布快照入口清单。
type ReleaseStaticManifest struct {
	Schema        string             `json:"schema"`
	ReleaseId     string             `json:"releaseId"`
	PluginId      string             `json:"pluginId"`
	Version       string             `json:"version"`
	RuntimeKind   ReleaseRuntimeKind `json:"runtimeKind"`
	AssetMetaUrl  string             `json:"assetMetaUrl,omitempty"`
	TemplateFiles map[string]string  `json:"templateFiles,omitempty"`
}

// 静态快照根目录。
func FactoryStaticRoot() string {
	root := strings.TrimSpace(setting.Config.App.FilePath.Factory)
	if root != "" {
		return root
	}
	if base := strings.TrimSpace(setting.Config.App.RuntimeRootPath); base != "" {
		return filepath.Join(base, "factory")
	}
	return filepath.Join("runtime", "factory")
}

// 站内静态访问路径。
func FactoryStaticURL(pathParts ...string) string {
	parts := append([]string{"static", "factory"}, pathParts...)
	return "/" + filepath.ToSlash(filepath.Join(parts...))
}

// 发布快照目录。
func ReleaseStaticDir(release Release) string {
	return filepath.Join(FactoryStaticRoot(), "releases", release.PluginId, fmt.Sprintf("%s-%d", release.Version, release.Id))
}

// 发布 manifest 访问路径。
func ReleaseStaticURL(release Release) string {
	return FactoryStaticURL("releases", release.PluginId, fmt.Sprintf("%s-%d", release.Version, release.Id), "release.json")
}

// 独立资产快照文件。
func AssetStaticPath(pluginId string, assetId int64) string {
	return filepath.Join(FactoryStaticRoot(), "assets", pluginId, fmt.Sprintf("%d.json", assetId))
}

// 独立资产访问路径。
func AssetStaticURL(pluginId string, assetId int64) string {
	return FactoryStaticURL("assets", pluginId, fmt.Sprintf("%d.json", assetId))
}

// 钱包地址不可逆索引。
func OwnerIndexKey(walletAddress string) string {
	normalized := strings.ToLower(strings.TrimSpace(walletAddress))
	sum := sha256.Sum256([]byte(factoryOwnerHashSalt + normalized))
	return hex.EncodeToString(sum[:])
}

// 持有人资产索引文件。
func OwnerIndexStaticPath(ownerKey string) string {
	return filepath.Join(FactoryStaticRoot(), "owners", ownerKey, "assets.json")
}

// 持有人资产索引访问路径。
func OwnerIndexStaticURL(ownerKey string) string {
	return FactoryStaticURL("owners", ownerKey, "assets.json")
}

// 持有人组合快照文件。
func OwnerCompositionStaticPath(ownerKey string) string {
	return filepath.Join(FactoryStaticRoot(), "owners", ownerKey, "composition.json")
}

// 持有人组合快照访问路径。
func OwnerCompositionStaticURL(ownerKey string) string {
	return FactoryStaticURL("owners", ownerKey, "composition.json")
}

// 从插件模板重建发布快照。
func EnsureReleaseStaticSnapshot(release Release) error {
	dir := ReleaseStaticDir(release)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	templateDir := filepath.Join("asset", "factory-templates", release.PluginId)
	templateFiles := map[string]string{}
	if entries, err := os.ReadDir(templateDir); err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			src := filepath.Join(templateDir, entry.Name())
			dst := filepath.Join(dir, entry.Name())
			if err := copyFile(src, dst); err != nil {
				return err
			}
			templateFiles[entry.Name()] = FactoryStaticURL("releases", release.PluginId, fmt.Sprintf("%s-%d", release.Version, release.Id), entry.Name())
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	manifest := ReleaseStaticManifest{
		Schema:        "senspace.factory.release.v1",
		ReleaseId:     fmt.Sprintf("%d", release.Id),
		PluginId:      release.PluginId,
		Version:       release.Version,
		RuntimeKind:   release.RuntimeKind,
		TemplateFiles: templateFiles,
	}
	if url := templateFiles["asset.meta.json"]; url != "" {
		manifest.AssetMetaUrl = url
	}
	return writeJSONAtomic(filepath.Join(dir, "release.json"), manifest)
}

// 原子写入 JSON。
func WriteJSONAtomic(path string, value any) error {
	return writeJSONAtomic(path, value)
}

// 写入临时文件后原子替换目标文件。
func writeJSONAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// 复制单个文件并创建目标目录。
func copyFile(src string, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

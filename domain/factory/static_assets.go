package factory

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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
	BasePrice     string             `json:"basePrice"`
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

// 发布快照临时目录。
func ReleaseStaticStagingDir(release Release) string {
	return filepath.Join(FactoryStaticRoot(), "releases", ".staging", release.PluginId, fmt.Sprintf("%s-%d", release.Version, release.Id))
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

// 标准 NFT metadata 文件。
func MetadataStaticPath(pluginId string, tokenId string) string {
	return filepath.Join(FactoryStaticRoot(), "metadata", pluginId, fmt.Sprintf("%s.json", tokenId))
}

// 标准 NFT metadata 访问路径。
func MetadataStaticURL(pluginId string, tokenId string) string {
	return FactoryStaticURL("metadata", pluginId, fmt.Sprintf("%s.json", tokenId))
}

// NFT 可验证 proof 文件。
func ProofStaticPath(pluginId string, tokenId string) string {
	return filepath.Join(FactoryStaticRoot(), "proofs", pluginId, fmt.Sprintf("%s.json", tokenId))
}

// NFT 可验证 proof 访问路径。
func ProofStaticURL(pluginId string, tokenId string) string {
	return FactoryStaticURL("proofs", pluginId, fmt.Sprintf("%s.json", tokenId))
}

// 发布后 item metadata 文件。
func ItemMetadataStaticPath(pluginId string, collectionKey string, itemId string) string {
	return filepath.Join(FactoryStaticRoot(), "metadata", pluginId, "by-item", collectionKey, fmt.Sprintf("%s.json", itemId))
}

// 发布后 item metadata 访问路径。
func ItemMetadataStaticURL(pluginId string, collectionKey string, itemId string) string {
	return FactoryStaticURL("metadata", pluginId, "by-item", collectionKey, fmt.Sprintf("%s.json", itemId))
}

// 发布后 item proof 文件。
func ItemProofStaticPath(pluginId string, collectionKey string, itemId string) string {
	return filepath.Join(FactoryStaticRoot(), "proofs", pluginId, "by-item", collectionKey, fmt.Sprintf("%s.json", itemId))
}

// 发布后 item proof 访问路径。
func ItemProofStaticURL(pluginId string, collectionKey string, itemId string) string {
	return FactoryStaticURL("proofs", pluginId, "by-item", collectionKey, fmt.Sprintf("%s.json", itemId))
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
	return buildReleaseStaticSnapshot(release, ReleaseStaticDir(release))
}

// 在 staging 目录中构建发布快照；调用方确认成功后再激活为正式快照。
func StageReleaseStaticSnapshot(release Release) (string, error) {
	dir := ReleaseStaticStagingDir(release)
	if err := os.RemoveAll(dir); err != nil {
		return "", err
	}
	if err := buildReleaseStaticSnapshot(release, dir); err != nil {
		_ = os.RemoveAll(dir)
		return "", err
	}
	return dir, nil
}

// 激活 staging 快照，并保留旧正式目录作为回滚备份。
func ActivateReleaseStaticSnapshot(release Release, stagingDir string) (string, error) {
	finalDir := ReleaseStaticDir(release)
	backupDir := finalDir + ".previous"
	if err := os.MkdirAll(filepath.Dir(finalDir), 0755); err != nil {
		return "", err
	}
	if err := os.RemoveAll(backupDir); err != nil {
		return "", err
	}

	hasFinal := false
	if info, err := os.Stat(finalDir); err == nil && info.IsDir() {
		hasFinal = true
		if err := os.Rename(finalDir, backupDir); err != nil {
			return "", err
		}
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	if err := os.Rename(stagingDir, finalDir); err != nil {
		if hasFinal {
			_ = os.Rename(backupDir, finalDir)
		}
		return "", err
	}
	if !hasFinal {
		backupDir = ""
	}
	return backupDir, nil
}

// 提交已激活的发布快照，清理回滚备份。
func CommitActivatedReleaseStaticSnapshot(backupDir string) error {
	if backupDir == "" {
		return nil
	}
	return os.RemoveAll(backupDir)
}

// 回滚已激活的发布快照。
func RollbackActivatedReleaseStaticSnapshot(release Release, backupDir string) error {
	finalDir := ReleaseStaticDir(release)
	if err := os.RemoveAll(finalDir); err != nil {
		return err
	}
	if backupDir == "" {
		return nil
	}
	if _, err := os.Stat(backupDir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.Rename(backupDir, finalDir)
}

func buildReleaseStaticSnapshot(release Release, dir string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	templateFiles := map[string]string{}
	if err := copyReleaseTemplateSnapshot(dir, release, templateFiles); err != nil {
		return err
	}

	basePrice := strings.TrimSpace(release.MintPrice)
	if basePrice == "" {
		basePrice = "0"
	}
	manifest := ReleaseStaticManifest{
		Schema:        "senspace.factory.release.v1",
		ReleaseId:     fmt.Sprintf("%d", release.Id),
		PluginId:      release.PluginId,
		Version:       release.Version,
		RuntimeKind:   release.RuntimeKind,
		BasePrice:     basePrice,
		TemplateFiles: templateFiles,
	}
	if url := templateFiles["asset.meta.json"]; url != "" {
		manifest.AssetMetaUrl = url
	}
	return writeJSONAtomic(filepath.Join(dir, "release.json"), manifest)
}

func copyReleaseTemplateSnapshot(dir string, release Release, templateFiles map[string]string) error {
	templateDir := filepath.Join("asset", "factory-templates", release.PluginId)
	if info, err := os.Stat(templateDir); err == nil && info.IsDir() {
		return copyJSONTree(templateDir, dir, "", release, templateFiles)
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}

	srcRoot := pluginSourceRoot(release.PluginId)
	if srcRoot == "" {
		return nil
	}
	assetMetaPath := filepath.Join(srcRoot, "asset.meta.json")
	if _, err := os.Stat(assetMetaPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return copyPluginSourceReleaseSnapshot(srcRoot, dir, release, templateFiles)
}

// 递归复制 JSON 模板文件，并写入发布 manifest 的模板文件索引。
func copyJSONTree(srcRoot string, dstRoot string, prefix string, release Release, templateFiles map[string]string) error {
	entries, err := os.ReadDir(srcRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		rel := name
		if prefix != "" {
			rel = filepath.Join(prefix, name)
		}
		src := filepath.Join(srcRoot, name)
		if entry.IsDir() {
			if err := copyJSONTree(src, dstRoot, rel, release, templateFiles); err != nil {
				return err
			}
			continue
		}
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		dst := filepath.Join(dstRoot, rel)
		if err := copyFile(src, dst); err != nil {
			return err
		}
		key := filepath.ToSlash(rel)
		templateFiles[key] = FactoryStaticURL(
			"releases",
			release.PluginId,
			fmt.Sprintf("%s-%d", release.Version, release.Id),
			key,
		)
	}
	return nil
}

// 从插件源码目录收集 asset.meta.json 和它引用的 JSON 模板文件。
func copyPluginSourceReleaseSnapshot(srcRoot string, dir string, release Release, templateFiles map[string]string) error {
	if err := copyFile(filepath.Join(srcRoot, "asset.meta.json"), filepath.Join(dir, "asset.meta.json")); err != nil {
		return err
	}
	templateFiles["asset.meta.json"] = FactoryStaticURL(
		"releases",
		release.PluginId,
		fmt.Sprintf("%s-%d", release.Version, release.Id),
		"asset.meta.json",
	)

	files, err := collectAssetMetaReferencedFiles(filepath.Join(srcRoot, "asset.meta.json"))
	if err != nil {
		return err
	}
	for _, rel := range files {
		if err := copyPluginSourceJSONFile(srcRoot, dir, rel, release, templateFiles); err != nil {
			return err
		}
	}
	return nil
}

type assetMetaReferenceTemplate struct {
	Collections []assetMetaReferenceCollection `json:"collections"`
}

type assetMetaReferenceCollection struct {
	MetadataRef string                     `json:"metadataRef"`
	TierConfig  map[string]json.RawMessage `json:"tierConfig"`
}

func collectAssetMetaReferencedFiles(path string) ([]string, error) {
	var template assetMetaReferenceTemplate
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &template); err != nil {
		return nil, err
	}
	seen := map[string]struct{}{}
	files := make([]string, 0, len(template.Collections))
	for _, collection := range template.Collections {
		ref := strings.TrimSpace(collection.MetadataRef)
		if ref == "" {
			continue
		}
		if len(collection.TierConfig) > 0 && strings.Contains(ref, "{tier}") {
			for tier := range collection.TierConfig {
				file := assetMetaReferenceFile(strings.ReplaceAll(ref, "{tier}", strings.TrimSpace(tier)))
				if file == "" {
					continue
				}
				if _, ok := seen[file]; ok {
					continue
				}
				seen[file] = struct{}{}
				files = append(files, file)
			}
			continue
		}
		file := assetMetaReferenceFile(ref)
		if file == "" {
			continue
		}
		if _, ok := seen[file]; ok {
			continue
		}
		seen[file] = struct{}{}
		files = append(files, file)
	}
	sort.Strings(files)
	return files, nil
}

func assetMetaReferenceFile(ref string) string {
	parts := strings.SplitN(strings.TrimSpace(ref), "#", 2)
	return strings.TrimSpace(parts[0])
}

func copyPluginSourceJSONFile(srcRoot string, dstRoot string, rel string, release Release, templateFiles map[string]string) error {
	src := firstExistingFile(pluginTemplateFileCandidates(srcRoot, rel))
	if src == "" {
		return fmt.Errorf("plugin template file not found: %s", rel)
	}
	dst := filepath.Join(dstRoot, rel)
	if err := copyFile(src, dst); err != nil {
		return err
	}
	key := filepath.ToSlash(rel)
	templateFiles[key] = FactoryStaticURL(
		"releases",
		release.PluginId,
		fmt.Sprintf("%s-%d", release.Version, release.Id),
		key,
	)
	return nil
}

// 返回插件模板文件的候选路径。
func pluginTemplateFileCandidates(srcRoot string, rel string) []string {
	candidates := []string{filepath.Join(srcRoot, rel)}
	entries, err := os.ReadDir(srcRoot)
	if err != nil {
		return candidates
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(entry.Name()))
		if !strings.Contains(name, "generator") {
			continue
		}
		candidates = append(candidates, filepath.Join(srcRoot, entry.Name(), rel))
	}
	return candidates
}

// 返回第一个存在的目录。
func firstExistingDir(candidates []string) string {
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}
	}
	return ""
}

// 返回第一个存在的文件。
func firstExistingFile(candidates []string) string {
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}

func pluginSourceRoot(pluginId string) string {
	return firstExistingDir([]string{
		filepath.Join("..", "senspace-web", "src", "components", "StarSky", "Desktop", "Plugins", pluginId),
		filepath.Join("senspace-web", "src", "components", "StarSky", "Desktop", "Plugins", pluginId),
	})
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
	if closeErr != nil {
		return closeErr
	}
	return nil
}

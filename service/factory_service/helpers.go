package factory_service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"math/big"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"senspace/domain"
	"senspace/domain/dev"
	"senspace/domain/factory"
	factoryvo "senspace/domain/factory/vo"
	"senspace/pkg/app/security"
	"senspace/pkg/setting"
	"senspace/pkg/util"

	"gorm.io/gorm"
)

const (
	defaultCategory  = "tool"
	defaultZeroPrice = "0"
)

// 数据库连接。
func db() (*gorm.DB, error) {
	if domain.Db == nil {
		return nil, fmt.Errorf("factory db not initialized")
	}
	return domain.Db, nil
}

// 解析字符串 ID。
func parseID(raw string, field string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, newParameterError(field + "不能为空")
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return 0, newParameterError(field + "无效")
	}
	return value, nil
}

// 清洗 manifest。
func normalizeManifest(m PluginManifest) PluginManifest {
	entry := strings.TrimSpace(m.Entry)
	if entry == "" {
		entry = "src/index.ts"
	}
	return PluginManifest{
		Name:        strings.TrimSpace(m.Name),
		Version:     strings.TrimSpace(m.Version),
		Entry:       entry,
		Description: strings.TrimSpace(m.Description),
	}
}

// 清洗标签。
func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{}, len(tags))
	result := make([]string, 0, len(tags))
	for _, tag := range tags {
		value := strings.TrimSpace(tag)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

// 清洗市场信息。
func normalizeMarket(m MutableMarketMetadata) MutableMarketMetadata {
	return MutableMarketMetadata{
		Summary:  strings.TrimSpace(m.Summary),
		Category: strings.TrimSpace(m.Category),
		Tags:     normalizeTags(m.Tags),
		CoverUrl: strings.TrimSpace(m.CoverUrl),
	}
}

// 清洗发布配置。
func normalizeReleasePayload(r ReleasePayload) ReleasePayload {
	r.MutableMarketMetadata = normalizeMarket(r.MutableMarketMetadata)
	r.MintPrice = strings.TrimSpace(r.MintPrice)
	r.UpgradePrice = strings.TrimSpace(r.UpgradePrice)
	if r.Category == "" {
		r.Category = defaultCategory
	}
	if r.UpgradePolicy == "" {
		r.UpgradePolicy = factory.ReleaseUpgradePolicyNone
	}
	return r
}

// 校验金额字符串。
func validateDecimalString(raw string, field string, allowEmpty bool) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		if allowEmpty {
			return "", nil
		}
		return "", newParameterError(field + "不能为空")
	}
	rat, ok := new(big.Rat).SetString(raw)
	if !ok {
		return "", newParameterError(field + "格式错误")
	}
	if rat.Sign() < 0 {
		return "", newParameterError(field + "不能为负数")
	}
	return normalizeDecimal(rat.FloatString(18)), nil
}

// 规范化金额。
func normalizeDecimal(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !strings.Contains(raw, ".") {
		return raw
	}
	raw = strings.TrimRight(raw, "0")
	raw = strings.TrimRight(raw, ".")
	if raw == "" {
		return defaultZeroPrice
	}
	return raw
}

// 比较版本号。
func compareVersions(v1, v2 string) int {
	if v1 == v2 {
		return 0
	}
	p1 := strings.Split(v1, ".")
	p2 := strings.Split(v2, ".")
	size := len(p1)
	if len(p2) > size {
		size = len(p2)
	}
	for i := 0; i < size; i++ {
		n1 := versionPart(p1, i)
		n2 := versionPart(p2, i)
		if n1 == n2 {
			continue
		}
		if n1 > n2 {
			return 1
		}
		return -1
	}
	return strings.Compare(v1, v2)
}

// 读取版本号片段。
func versionPart(parts []string, index int) int {
	if index >= len(parts) {
		return 0
	}
	value, err := strconv.Atoi(parts[index])
	if err != nil {
		return 0
	}
	return value
}

// 主版本号。
func majorVersion(version string) int {
	return versionPart(strings.Split(version, "."), 0)
}

// 判断是否允许升级。
func releaseAllowsUpgrade(fromVersion string, target factory.Release) (factory.UpgradeType, string, bool) {
	switch target.UpgradePolicy {
	case factory.ReleaseUpgradePolicyFree:
		return factory.UpgradeTypeFree, defaultZeroPrice, true
	case factory.ReleaseUpgradePolicyPaid:
		return factory.UpgradeTypePaid, zeroIfEmpty(target.UpgradePrice), true
	case factory.ReleaseUpgradePolicyMajorPaid:
		if majorVersion(fromVersion) == majorVersion(target.Version) {
			return factory.UpgradeTypeFree, defaultZeroPrice, true
		}
		return factory.UpgradeTypePaid, zeroIfEmpty(target.UpgradePrice), true
	default:
		return "", "", false
	}
}

// 空值转 0。
func zeroIfEmpty(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return defaultZeroPrice
	}
	return normalizeDecimal(value)
}

// 默认运行来源。
func defaultRuntimeKind(kind factory.ReleaseRuntimeKind) factory.ReleaseRuntimeKind {
	switch kind {
	case factory.ReleaseRuntimeKindBuiltin, factory.ReleaseRuntimeKindBook:
		return kind
	default:
		return factory.ReleaseRuntimeKindArtifact
	}
}

// 插件源码目录。
func getPluginRoot(pluginId string) string {
	return filepath.Join(setting.Config.App.FilePath.Plugin, pluginId)
}

// 最新版本目录。
func resolveLatestPluginVersionRoot(pluginId string) (string, string, error) {
	root := getPluginRoot(pluginId)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", newNotFoundError("插件源码目录不存在")
		}
		return "", "", err
	}

	latestVersion := ""
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		version := entry.Name()
		if latestVersion == "" || compareVersions(version, latestVersion) > 0 {
			latestVersion = version
		}
	}
	if latestVersion == "" {
		return "", "", newNotFoundError("插件源码版本不存在")
	}
	return latestVersion, filepath.Join(root, latestVersion), nil
}

// 读取 manifest。
func loadManifestFromDir(versionRoot string) (PluginManifest, error) {
	manifestPath := filepath.Join(versionRoot, "manifest.json")
	bytes, err := os.ReadFile(manifestPath)
	if err != nil {
		if os.IsNotExist(err) {
			return PluginManifest{}, newParameterError("插件目录缺少manifest.json")
		}
		return PluginManifest{}, err
	}

	var parsed struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		Entry       string `json:"entry"`
		Main        string `json:"main"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(bytes, &parsed); err != nil {
		return PluginManifest{}, newParameterError("manifest.json格式错误")
	}

	entry := strings.TrimSpace(parsed.Entry)
	if entry == "" {
		entry = strings.TrimSpace(parsed.Main)
	}
	if entry == "" {
		entry = "src/index.ts"
	}

	return normalizeManifest(PluginManifest{
		Name:        parsed.Name,
		Version:     parsed.Version,
		Entry:       entry,
		Description: parsed.Description,
	}), nil
}

// 校验发布请求。
func validatePublishRequest(req PublishRequest) (PublishRequest, error) {
	req.PluginId = strings.TrimSpace(req.PluginId)
	req.Manifest = normalizeManifest(req.Manifest)
	req.Release = normalizeReleasePayload(req.Release)

	if req.PluginId == "" {
		return req, newParameterError("pluginId不能为空")
	}
	if req.Manifest.Name == "" {
		return req, newParameterError("插件名称不能为空")
	}
	if req.Manifest.Version == "" {
		return req, newParameterError("插件版本不能为空")
	}
	if req.Manifest.Entry == "" {
		return req, newParameterError("插件入口不能为空")
	}
	if req.Release.Summary == "" {
		return req, newParameterError("市场摘要不能为空")
	}
	if req.Release.Category == "" {
		return req, newParameterError("分类不能为空")
	}
	if req.Release.TotalSupply <= 0 {
		return req, newParameterError("总发行量必须大于0")
	}
	if req.Release.MintPer <= 0 {
		return req, newParameterError("单次最大铸造量必须大于0")
	}
	if req.Release.MintPer > req.Release.TotalSupply {
		return req, newParameterError("单次最大铸造量不能超过总发行量")
	}

	mintPrice, err := validateDecimalString(req.Release.MintPrice, "铸造价格", false)
	if err != nil {
		return req, err
	}
	req.Release.MintPrice = mintPrice

	switch req.Release.UpgradePolicy {
	case "", factory.ReleaseUpgradePolicyNone:
		req.Release.UpgradePolicy = factory.ReleaseUpgradePolicyNone
		req.Release.UpgradePrice = ""
	case factory.ReleaseUpgradePolicyFree:
		req.Release.UpgradePrice = defaultZeroPrice
	case factory.ReleaseUpgradePolicyPaid, factory.ReleaseUpgradePolicyMajorPaid:
		upgradePrice, err := validateDecimalString(req.Release.UpgradePrice, "升级价格", false)
		if err != nil {
			return req, err
		}
		req.Release.UpgradePrice = upgradePrice
	default:
		return req, newParameterError("升级策略不支持")
	}

	return req, nil
}

// 校验插件归属。
func ensurePluginExists(tx *gorm.DB, pluginId string, userId uint64) (*dev.Plugin, error) {
	var plugin dev.Plugin
	if err := tx.Where("id = ?", pluginId).First(&plugin).Error; err == nil {
		if plugin.CreatedBy != 0 && plugin.CreatedBy != userId {
			return nil, newForbiddenError("无权发布他人的插件")
		}
		return &plugin, nil
	}

	if _, err := os.Stat(getPluginRoot(pluginId)); err == nil {
		return nil, nil
	} else if os.IsNotExist(err) {
		return nil, newNotFoundError("插件不存在")
	} else {
		return nil, err
	}
}

// 比对 manifest。
func compareManifest(expected PluginManifest, actual PluginManifest) error {
	if expected != actual {
		return newConflictError("请求manifest与插件仓库快照不一致")
	}
	return nil
}

// 生成哈希。
func buildHashes(versionRoot string, manifest PluginManifest) (string, string, error) {
	files := make([]string, 0, 16)
	if err := filepath.WalkDir(versionRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(versionRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if rel == "publish.json" {
			return nil
		}
		files = append(files, rel)
		return nil
	}); err != nil {
		return "", "", err
	}
	sort.Strings(files)

	sourceHasher := sha256.New()
	bundleHasher := sha256.New()
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return "", "", err
	}
	sourceHasher.Write(manifestBytes)

	for _, rel := range files {
		bytes, err := os.ReadFile(filepath.Join(versionRoot, filepath.FromSlash(rel)))
		if err != nil {
			return "", "", err
		}
		sourceHasher.Write([]byte(rel))
		sourceHasher.Write([]byte{'\n'})
		sourceHasher.Write(bytes)
	}

	sourceHash := "sha256:" + hex.EncodeToString(sourceHasher.Sum(nil))
	bundleHasher.Write([]byte(sourceHash))
	bundleHasher.Write([]byte{'\n'})
	bundleHasher.Write(manifestBytes)
	bundleHash := "sha256:" + hex.EncodeToString(bundleHasher.Sum(nil))
	return sourceHash, bundleHash, nil
}

// 时间指针。
func nowPtr(t time.Time) *time.Time {
	return &t
}

// 转为 manifest 快照。
func toManifestSnapshot(m PluginManifest) factory.PluginManifestSnapshot {
	return factory.PluginManifestSnapshot{
		Name:        m.Name,
		Version:     m.Version,
		Entry:       m.Entry,
		Description: m.Description,
	}
}

// 转为作者快照。
func toAuthorSnapshot(author security.JwtUser) factory.AuthorSnapshot {
	return factory.AuthorSnapshot{
		Id:     strconv.FormatUint(author.Id, 10),
		Name:   strings.TrimSpace(author.Nickname),
		Avatar: strings.TrimSpace(author.Avatar),
	}
}

// 映射作者信息。
func mapAuthor(snapshot factory.AuthorSnapshot, authorId uint64) AuthorProfile {
	id := strings.TrimSpace(snapshot.Id)
	if id == "" {
		id = strconv.FormatUint(authorId, 10)
	}

	return AuthorProfile{
		Id:     id,
		Name:   strings.TrimSpace(snapshot.Name),
		Avatar: strings.TrimSpace(snapshot.Avatar),
	}
}

// 映射价格历史。
func mapPriceHistory(record factory.ReleasePriceHistory) PriceHistoryRecord {
	return PriceHistoryRecord{
		Id:                strconv.FormatInt(record.Id, 10),
		ReleaseId:         strconv.FormatInt(record.ReleaseId, 10),
		PreviousMintPrice: trimDecimal(record.PreviousMintPrice),
		NextMintPrice:     trimDecimal(record.NextMintPrice),
		Reason:            record.Reason,
		ChangedBy:         record.ChangedBy,
		ChangedAt:         record.ChangedAt.Format(time.RFC3339Nano),
	}
}

// 映射发布记录。
func mapRelease(record factory.Release) PublishRecord {
	return PublishRecord{
		Id:             strconv.FormatInt(record.Id, 10),
		PluginId:       record.PluginId,
		Author:         mapAuthor(record.AuthorSnapshot, record.AuthorId),
		Name:           record.Name,
		Version:        record.Version,
		Status:         record.Status,
		ReviewStatus:   record.ReviewStatus,
		CurrentRelease: record.CurrentRelease,
		TotalSupply:    record.TotalSupply,
		MintedCount:    record.MintedCount,
		MintPrice:      trimDecimal(record.MintPrice),
		Summary:        record.Summary,
		Category:       record.Category,
		Tags:           []string(record.Tags),
		CoverUrl:       record.CoverUrl,
		SourceHash:     record.SourceHash,
		BundleHash:     record.BundleHash,
		Integrity:      record.Integrity,
		BuildStatus:    defaultBuildStatus(record.BuildStatus),
		BuildError:     record.BuildError,
		BuiltAt:        formatTime(record.BuiltAt),
		RuntimeKind:    defaultRuntimeKind(record.RuntimeKind),
		ReleaseUrl:     factory.ReleaseStaticURL(record),
		UpgradePolicy:  defaultUpgradePolicy(record.UpgradePolicy),
		UpgradePrice:   displayUpgradePrice(record.UpgradePolicy, record.UpgradePrice),
		PublishedAt:    formatTime(record.PublishedAt),
		PausedAt:       formatTime(record.PausedAt),
		ClosedAt:       formatTime(record.ClosedAt),
		UpdatedAt:      record.UpdatedAt.Format(time.RFC3339Nano),
	}
}

// 映射发布详情。
func mapReleaseDetail(record factory.Release, history []factory.ReleasePriceHistory) PublishDetail {
	priceHistory := make([]PriceHistoryRecord, 0, len(history))
	for _, item := range history {
		priceHistory = append(priceHistory, mapPriceHistory(item))
	}

	return PublishDetail{
		PublishRecord: mapRelease(record),
		ManifestSnapshot: PluginManifest{
			Name:        record.ManifestSnapshot.Name,
			Version:     record.ManifestSnapshot.Version,
			Entry:       record.ManifestSnapshot.Entry,
			Description: record.ManifestSnapshot.Description,
		},
		PriceHistory: priceHistory,
	}
}

// 映射资产视图。
func mapOwnership(
	record factory.UserOwnership,
	mintedVersion string,
	effectiveVersion string,
	latestRelease *factory.Release,
) factoryvo.UserPluginOwnershipView {
	view := factoryvo.UserPluginOwnershipView{
		Id:                     strconv.FormatInt(record.Id, 10),
		UserId:                 strconv.FormatUint(record.UserId, 10),
		PluginId:               record.PluginId,
		MintedReleaseId:        strconv.FormatInt(record.MintedReleaseId, 10),
		EffectiveReleaseId:     strconv.FormatInt(record.EffectiveReleaseId, 10),
		CreatedAt:              record.CreatedAt.Format(time.RFC3339Nano),
		UpdatedAt:              record.UpdatedAt.Format(time.RFC3339Nano),
		UpgradedAt:             formatTime(record.UpgradedAt),
		MintedVersion:          mintedVersion,
		EffectiveVersion:       effectiveVersion,
		UpgradeState:           factory.OwnershipUpgradeStateUpToDate,
		LatestAvailableVersion: "",
	}

	if latestRelease == nil {
		return view
	}

	upgradeType, upgradePrice, allowed := releaseAllowsUpgrade(effectiveVersion, *latestRelease)
	if !allowed || compareVersions(latestRelease.Version, effectiveVersion) <= 0 {
		return view
	}

	view.LatestAvailableReleaseId = strconv.FormatInt(latestRelease.Id, 10)
	view.LatestAvailableVersion = latestRelease.Version
	if upgradeType == factory.UpgradeTypePaid {
		view.UpgradeState = factory.OwnershipUpgradeStateUpgradeRequired
		view.UpgradePrice = trimDecimal(upgradePrice)
	} else {
		view.UpgradeState = factory.OwnershipUpgradeStateUpgradable
		view.UpgradePrice = defaultZeroPrice
	}
	return view
}

// 映射升级记录。
func mapUpgradeRecord(record factory.UpgradeRecord) UpgradeRecord {
	return UpgradeRecord{
		Id:            strconv.FormatInt(record.Id, 10),
		OwnershipId:   strconv.FormatInt(record.OwnershipId, 10),
		UserId:        strconv.FormatUint(record.UserId, 10),
		PluginId:      record.PluginId,
		FromReleaseId: strconv.FormatInt(record.FromReleaseId, 10),
		ToReleaseId:   strconv.FormatInt(record.ToReleaseId, 10),
		UpgradeType:   record.UpgradeType,
		PaidAmount:    trimDecimal(record.PaidAmount),
		UpgradedAt:    record.UpgradedAt.Format(time.RFC3339Nano),
	}
}

// 格式化时间。
func formatTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format(time.RFC3339Nano)
}

// 整理金额显示。
func trimDecimal(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	rat, ok := new(big.Rat).SetString(value)
	if !ok {
		return value
	}
	return normalizeDecimal(rat.FloatString(18))
}

// 补默认构建状态。
func defaultBuildStatus(status factory.BuildStatus) factory.BuildStatus {
	if status == "" {
		return factory.BuildStatusPending
	}
	return status
}

// 补默认升级策略。
func defaultUpgradePolicy(policy factory.ReleaseUpgradePolicy) factory.ReleaseUpgradePolicy {
	if policy == "" {
		return factory.ReleaseUpgradePolicyNone
	}
	return policy
}

// 处理升级价格展示。
func displayUpgradePrice(policy factory.ReleaseUpgradePolicy, value string) string {
	switch defaultUpgradePolicy(policy) {
	case factory.ReleaseUpgradePolicyPaid, factory.ReleaseUpgradePolicyMajorPaid:
		return trimDecimal(value)
	case factory.ReleaseUpgradePolicyFree:
		return defaultZeroPrice
	default:
		return ""
	}
}

// 生成主键。
func generateID() int64 {
	if util.IdWorker == nil {
		util.Setup()
	}
	return util.IdWorker.Generate().Int64()
}

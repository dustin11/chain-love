package factory_service

import (
	crand "crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"senspace/domain/factory"
	"senspace/pkg/app/security"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 对应 asset.meta.json。
type assetValueTemplate struct {
	Schema    string `json:"schema"`
	PluginId  string `json:"pluginId"`
	BasePrice string `json:"basePrice"`
	// 数组顺序即前端展示顺序。
	Collections []assetValueCollection `json:"collections"`
}

// 可铸造模板集合。
type assetValueCollection struct {
	Label     string            `json:"label"`
	Key       string            `json:"key"`
	AssetKind factory.AssetKind `json:"assetKind"`
	// 例如 defaultWaterMeta.json#fish。
	Ref string `json:"ref"`
	// 模板项中用于区分等级的字段。
	TierField string `json:"tierField"`
	// 单等级上限，0 表示不限。
	MaxTierCount    int64             `json:"maxTierCount"`
	UnitPrice       string            `json:"unitPrice"`
	UnitPriceByTier map[string]string `json:"unitPriceByTier"`
}

// 计价后的铸造明细。
type mintSelectionResult struct {
	Items        []mintSelectionItem
	ExpectedPaid string
	TotalCount   int64
}

// 单个待生成 NFT。
type mintSelectionItem struct {
	AssetKind    factory.AssetKind
	TemplateRef  string
	TemplateFile string
	RefField     string
	TemplateId   string
}

// 静态目录中的单个 NFT 快照。
type mintedFactoryAsset struct {
	Schema       string                     `json:"schema"`
	AssetId      string                     `json:"assetId"`
	PluginId     string                     `json:"pluginId"`
	ReleaseId    string                     `json:"releaseId"`
	Version      string                     `json:"version"`
	RuntimeKind  factory.ReleaseRuntimeKind `json:"runtimeKind"`
	AssetKind    factory.AssetKind          `json:"assetKind"`
	TemplateRef  string                     `json:"templateRef"`
	TemplateId   string                     `json:"templateId"`
	ReleaseUrl   string                     `json:"releaseUrl"`
	MintRecordId string                     `json:"mintRecordId"`
	MintedAt     string                     `json:"mintedAt"`
	TemplateRefs map[string]string          `json:"templateRefs,omitempty"`
}

// 钱包地址对应的可重建资产索引。
type ownerFactoryAssetIndex struct {
	Schema    string                   `json:"schema"`
	OwnerKey  string                   `json:"ownerKey"`
	UpdatedAt string                   `json:"updatedAt"`
	Assets    []ownerFactoryAssetEntry `json:"assets"`
}

// 索引中的资产摘要。
type ownerFactoryAssetEntry struct {
	AssetId     string                     `json:"assetId"`
	PluginId    string                     `json:"pluginId"`
	ReleaseId   string                     `json:"releaseId"`
	Version     string                     `json:"version"`
	RuntimeKind factory.ReleaseRuntimeKind `json:"runtimeKind"`
	AssetKind   factory.AssetKind          `json:"assetKind"`
	TemplateRef string                     `json:"templateRef"`
	TemplateId  string                     `json:"templateId"`
	AssetUrl    string                     `json:"assetUrl"`
	ReleaseUrl  string                     `json:"releaseUrl"`
	MintedAt    string                     `json:"mintedAt"`
}

// 钱包地址对应的组合快照。
type ownerFactoryAssetComposition struct {
	Schema    string                             `json:"schema"`
	OwnerKey  string                             `json:"ownerKey"`
	UpdatedAt string                             `json:"updatedAt"`
	Relations []ownerFactoryAssetCompositionEdge `json:"relations"`
}

// NFT 组合边。
type ownerFactoryAssetCompositionEdge struct {
	Id            string `json:"id"`
	RelationType  string `json:"relationType"`
	SourceAssetId string `json:"sourceAssetId"`
	TargetAssetId string `json:"targetAssetId"`
	Metadata      any    `json:"metadata,omitempty"`
}

// 按发布快照生成独立 NFT，并用 DB 保持权威状态。
func MintReleaseAsset(user security.JwtUser, releaseIdRaw string, req MintAssetRequest) (*MintAssetResponse, error) {
	releaseId, err := parseID(releaseIdRaw, "发布记录ID")
	if err != nil {
		return nil, err
	}
	if user.Id == 0 {
		return nil, newParameterError("用户ID不能为空")
	}
	walletAddress := strings.TrimSpace(user.Addr)
	if walletAddress == "" {
		return nil, newParameterError("钱包地址不能为空")
	}

	totalPaid, err := validateDecimalString(req.TotalPaid, "支付总额", false)
	if err != nil {
		return nil, err
	}
	req.TotalPaid = totalPaid

	tx, err := db()
	if err != nil {
		return nil, err
	}

	var response MintAssetResponse
	var ownerKey string
	err = tx.Transaction(func(tx *gorm.DB) error {
		var release factory.Release
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&release, "id = ?", releaseId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return newNotFoundError("发布记录不存在")
			}
			return err
		}
		if release.Status != factory.ReleaseStatusPublished {
			return newConflictError("当前发布记录不可铸造")
		}
		if defaultBuildStatus(release.BuildStatus) != factory.BuildStatusReady {
			return newConflictError("当前发布记录构建未完成")
		}
		if err := factory.EnsureReleaseStaticSnapshot(release); err != nil {
			return err
		}

		valueTemplate, err := loadReleaseMintTemplate(release)
		if err != nil {
			return err
		}
		mintSelection, err := resolveMintSelection(release, valueTemplate, req.Inputs)
		if err != nil {
			return err
		}
		if release.MintedCount+mintSelection.TotalCount > release.TotalSupply {
			return newConflictError("铸造数量超过可发行数量")
		}
		if mintSelection.ExpectedPaid != req.TotalPaid {
			return newParameterError(fmt.Sprintf("支付总额应为 %s", mintSelection.ExpectedPaid))
		}

		ownerKey = factory.OwnerIndexKey(walletAddress)
		record := factory.MintRecord{
			Id:            generateID(),
			PluginId:      release.PluginId,
			ReleaseId:     release.Id,
			UserId:        user.Id,
			WalletAddress: walletAddress,
			Quantity:      mintSelection.TotalCount,
			TotalPaid:     req.TotalPaid,
			ChainId:       req.ChainId,
			TxHash:        strings.TrimSpace(req.TxHash),
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		var createdAssets []factory.Asset
		for _, selected := range mintSelection.Items {
			asset := factory.Asset{
				Id:           generateID(),
				PluginId:     release.PluginId,
				ReleaseId:    release.Id,
				Version:      release.Version,
				RuntimeKind:  defaultRuntimeKind(release.RuntimeKind),
				AssetKind:    selected.AssetKind,
				TemplateRef:  selected.TemplateRef,
				TemplateId:   selected.TemplateId,
				OwnerAddress: walletAddress,
				OwnerKey:     ownerKey,
				MintRecordId: record.Id,
				ChainId:      req.ChainId,
				Status:       factory.AssetStatusActive,
			}
			if err := tx.Create(&asset).Error; err != nil {
				return err
			}
			createdAssets = append(createdAssets, asset)
		}

		if err := createDefaultMintRelations(tx, ownerKey, createdAssets); err != nil {
			return err
		}

		release.MintedCount += mintSelection.TotalCount
		if release.MintedCount >= release.TotalSupply {
			release.Status = factory.ReleaseStatusSoldOut
		}
		if err := tx.Save(&release).Error; err != nil {
			return err
		}

		if err := ensurePluginOwnership(tx, user.Id, release); err != nil {
			return err
		}

		response = MintAssetResponse{
			Assets:              make([]MintAssetResponseAsset, 0, len(createdAssets)),
			OwnerIndexUrl:       factory.OwnerIndexStaticURL(ownerKey),
			OwnerCompositionUrl: factory.OwnerCompositionStaticURL(ownerKey),
			TotalPaid:           req.TotalPaid,
		}
		for _, asset := range createdAssets {
			response.Assets = append(response.Assets, MintAssetResponseAsset{
				AssetId:   strconv.FormatInt(asset.Id, 10),
				AssetKind: asset.AssetKind,
				AssetUrl:  factory.AssetStaticURL(asset.PluginId, asset.Id),
			})
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := rebuildOwnerFactorySnapshots(ownerKey); err != nil {
		return nil, err
	}
	return &response, nil
}

// 确保用户拥有该插件的权益记录。
func ensurePluginOwnership(tx *gorm.DB, userId uint64, release factory.Release) error {
	var ownership factory.UserOwnership
	err := tx.Where("user_id = ? AND plugin_id = ?", userId, release.PluginId).First(&ownership).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err == nil {
		return nil
	}
	ownership = factory.UserOwnership{
		Id:                 generateID(),
		UserId:             userId,
		PluginId:           release.PluginId,
		MintedReleaseId:    release.Id,
		EffectiveReleaseId: release.Id,
	}
	return tx.Create(&ownership).Error
}

// 读取发布快照中的铸造模板。
func loadReleaseMintTemplate(release factory.Release) (assetValueTemplate, error) {
	dir := factory.ReleaseStaticDir(release)
	var valueTemplate assetValueTemplate
	if err := readJSONFile(filepath.Join(dir, "asset.meta.json"), &valueTemplate); err != nil {
		return valueTemplate, err
	}
	return valueTemplate, nil
}

// 读取 JSON 文件到目标结构。
func readJSONFile(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

// 校验铸造输入并计算待生成资产。
func resolveMintSelection(
	release factory.Release,
	valueTemplate assetValueTemplate,
	inputs map[string]map[string]int64,
) (mintSelectionResult, error) {
	total := new(big.Rat)
	base, err := parsePriceRat(valueTemplate.BasePrice, "基础价格")
	if err != nil {
		return mintSelectionResult{}, err
	}
	total.Add(total, base)

	var result mintSelectionResult
	for _, collection := range valueTemplate.Collections {
		ref := strings.TrimSpace(collection.Ref)
		if ref == "" {
			continue
		}
		counts := inputs[ref]
		if len(counts) == 0 {
			continue
		}
		assetKind, err := resolveCollectionAssetKind(collection)
		if err != nil {
			return mintSelectionResult{}, err
		}
		templateFile, refField, err := parseAssetRef(ref)
		if err != nil {
			return mintSelectionResult{}, err
		}
		templateItems, err := loadTemplateItems(filepath.Join(factory.ReleaseStaticDir(release), templateFile), refField)
		if err != nil {
			return mintSelectionResult{}, err
		}

		if len(collection.UnitPriceByTier) > 0 {
			if err := appendTieredSelection(&result, total, collection, assetKind, templateFile, refField, templateItems, counts); err != nil {
				return mintSelectionResult{}, err
			}
			continue
		}
		if err := appendFlatSelection(&result, total, collection, assetKind, templateFile, refField, templateItems, counts); err != nil {
			return mintSelectionResult{}, err
		}
	}
	if result.TotalCount <= 0 {
		return mintSelectionResult{}, newParameterError("请选择要铸造的资产数量")
	}
	result.ExpectedPaid = normalizeDecimal(total.FloatString(18))
	return result, nil
}

// 返回集合的展示名称。
func collectionDisplayName(collection assetValueCollection) string {
	label := strings.TrimSpace(collection.Label)
	if label != "" {
		return label
	}
	return strings.TrimSpace(collection.Ref)
}

// 校验并返回集合资产类型。
func resolveCollectionAssetKind(collection assetValueCollection) (factory.AssetKind, error) {
	switch collection.AssetKind {
	case factory.AssetKindTank, factory.AssetKindFish:
		return collection.AssetKind, nil
	default:
		return "", newParameterError(collectionDisplayName(collection) + "缺少资产类型")
	}
}

// 追加按等级计价的资产选择项。
func appendTieredSelection(
	result *mintSelectionResult,
	total *big.Rat,
	collection assetValueCollection,
	assetKind factory.AssetKind,
	templateFile string,
	refField string,
	templateItems []map[string]any,
	counts map[string]int64,
) error {
	tierField := strings.TrimSpace(collection.TierField)
	if tierField == "" {
		tierField = "tier"
	}

	for tier, count := range counts {
		tier = strings.TrimSpace(tier)
		if tier == "" || count <= 0 {
			continue
		}
		unitRaw, ok := collection.UnitPriceByTier[tier]
		if !ok {
			return newParameterError("未知等级: " + tier)
		}
		if collection.MaxTierCount > 0 && count > collection.MaxTierCount {
			return newParameterError(fmt.Sprintf("%s等级 %s 不能超过 %d", collectionDisplayName(collection), tier, collection.MaxTierCount))
		}
		unit, err := parsePriceRat(unitRaw, tier+" 单价")
		if err != nil {
			return err
		}
		total.Add(total, new(big.Rat).Mul(unit, big.NewRat(count, 1)))
		ids, err := pickTemplateIdsByField(templateItems, tierField, tier, count)
		if err != nil {
			return err
		}
		appendSelectedItems(result, assetKind, collection.Ref, templateFile, refField, ids)
	}
	return nil
}

// 追加不分等级计价的资产选择项。
func appendFlatSelection(
	result *mintSelectionResult,
	total *big.Rat,
	collection assetValueCollection,
	assetKind factory.AssetKind,
	templateFile string,
	refField string,
	templateItems []map[string]any,
	counts map[string]int64,
) error {
	var count int64
	for _, itemCount := range counts {
		if itemCount > 0 {
			count += itemCount
		}
	}
	if count <= 0 {
		return nil
	}
	unit, err := parsePriceRat(collection.UnitPrice, collectionDisplayName(collection)+"单价")
	if err != nil {
		return err
	}
	total.Add(total, new(big.Rat).Mul(unit, big.NewRat(count, 1)))
	ids, err := pickAnyTemplateIds(templateItems, count)
	if err != nil {
		return err
	}
	appendSelectedItems(result, assetKind, collection.Ref, templateFile, refField, ids)
	return nil
}

// 将模板 ID 追加为待铸造项。
func appendSelectedItems(
	result *mintSelectionResult,
	assetKind factory.AssetKind,
	templateRef string,
	templateFile string,
	refField string,
	ids []string,
) {
	for _, id := range ids {
		result.Items = append(result.Items, mintSelectionItem{
			AssetKind:    assetKind,
			TemplateRef:  strings.TrimSpace(templateRef),
			TemplateFile: templateFile,
			RefField:     refField,
			TemplateId:   id,
		})
		result.TotalCount++
	}
}

// 解析 file.json#field 格式的模板引用。
func parseAssetRef(ref string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(ref), "#", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", newParameterError("资产模板 ref 格式应为 file.json#field")
	}
	file := filepath.Clean(strings.TrimSpace(parts[0]))
	if filepath.IsAbs(file) || file == "." || strings.HasPrefix(file, ".."+string(filepath.Separator)) || file == ".." {
		return "", "", newParameterError("资产模板 ref 文件路径不合法")
	}
	return file, strings.TrimSpace(parts[1]), nil
}

// 读取模板文件中的数组字段。
func loadTemplateItems(path string, field string) ([]map[string]any, error) {
	var root map[string]json.RawMessage
	if err := readJSONFile(path, &root); err != nil {
		return nil, err
	}
	raw, ok := root[field]
	if !ok {
		return nil, newParameterError("资产模板缺少字段: " + field)
	}
	var items []map[string]any
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, err
	}
	return items, nil
}

// 解析价格字符串为有理数。
func parsePriceRat(raw string, field string) (*big.Rat, error) {
	normalized, err := validateDecimalString(raw, field, false)
	if err != nil {
		return nil, err
	}
	rat, ok := new(big.Rat).SetString(normalized)
	if !ok {
		return nil, newParameterError(field + "格式错误")
	}
	return rat, nil
}

// 按字段值随机选择模板 ID。
func pickTemplateIdsByField(templates []map[string]any, field string, value string, count int64) ([]string, error) {
	candidates := make([]string, 0, len(templates))
	for _, item := range templates {
		id, _ := item["id"].(string)
		if strings.TrimSpace(id) != "" && fmt.Sprint(item[field]) == value {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil, newParameterError("没有可用资产模板: " + value)
	}
	return pickTemplateIdsFromCandidates(candidates, count)
}

// 从所有模板中随机选择 ID。
func pickAnyTemplateIds(templates []map[string]any, count int64) ([]string, error) {
	candidates := make([]string, 0, len(templates))
	for _, item := range templates {
		id, _ := item["id"].(string)
		if strings.TrimSpace(id) != "" {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil, newParameterError("没有可用资产模板")
	}
	return pickTemplateIdsFromCandidates(candidates, count)
}

// 从候选 ID 中随机抽取指定数量。
func pickTemplateIdsFromCandidates(candidates []string, count int64) ([]string, error) {
	result := make([]string, 0, count)
	for i := int64(0); i < count; i++ {
		index, err := secureRandomIndex(len(candidates))
		if err != nil {
			return nil, err
		}
		result = append(result, candidates[index])
	}
	return result, nil
}

// 返回安全随机下标。
func secureRandomIndex(max int) (int, error) {
	if max <= 0 {
		return 0, newParameterError("随机候选为空")
	}
	value, err := crand.Int(crand.Reader, big.NewInt(int64(max)))
	if err != nil {
		return 0, err
	}
	return int(value.Int64()), nil
}

// 为默认鱼缸和鱼创建包含关系。
func createDefaultMintRelations(tx *gorm.DB, ownerKey string, assets []factory.Asset) error {
	tankIDs := make([]int64, 0)
	for _, asset := range assets {
		if asset.AssetKind == factory.AssetKindTank {
			tankIDs = append(tankIDs, asset.Id)
		}
	}
	if len(tankIDs) == 0 {
		return nil
	}

	var fishIndex int
	for _, asset := range assets {
		if asset.AssetKind != factory.AssetKindFish {
			continue
		}
		relation := factory.AssetRelation{
			Id:            generateID(),
			OwnerKey:      ownerKey,
			RelationType:  "contains",
			SourceAssetId: tankIDs[fishIndex%len(tankIDs)],
			TargetAssetId: asset.Id,
			MetadataJson:  "{}",
			Status:        factory.AssetRelationStatusActive,
		}
		if err := tx.Create(&relation).Error; err != nil {
			return err
		}
		fishIndex++
	}
	return nil
}

// 重建持有人的资产索引和组合快照。
func rebuildOwnerFactorySnapshots(ownerKey string) error {
	tx, err := db()
	if err != nil {
		return err
	}

	var assets []factory.Asset
	if err := tx.
		Where("owner_key = ? AND status = ?", ownerKey, factory.AssetStatusActive).
		Order("created_at DESC").
		Find(&assets).Error; err != nil {
		return err
	}

	var relations []factory.AssetRelation
	if err := tx.
		Where("owner_key = ? AND status = ?", ownerKey, factory.AssetRelationStatusActive).
		Order("created_at ASC").
		Find(&relations).Error; err != nil {
		return err
	}

	now := time.Now().Format(time.RFC3339Nano)
	index := ownerFactoryAssetIndex{
		Schema:    "senspace.factory.owner-assets.v2",
		OwnerKey:  ownerKey,
		UpdatedAt: now,
		Assets:    make([]ownerFactoryAssetEntry, 0, len(assets)),
	}
	for _, asset := range assets {
		if err := writeFactoryAssetSnapshot(asset); err != nil {
			return err
		}
		index.Assets = append(index.Assets, ownerFactoryAssetEntry{
			AssetId:     strconv.FormatInt(asset.Id, 10),
			PluginId:    asset.PluginId,
			ReleaseId:   strconv.FormatInt(asset.ReleaseId, 10),
			Version:     asset.Version,
			RuntimeKind: defaultRuntimeKind(asset.RuntimeKind),
			AssetKind:   asset.AssetKind,
			TemplateRef: asset.TemplateRef,
			TemplateId:  asset.TemplateId,
			AssetUrl:    factory.AssetStaticURL(asset.PluginId, asset.Id),
			ReleaseUrl:  releaseStaticURLFromAsset(asset),
			MintedAt:    asset.CreatedAt.Format(time.RFC3339Nano),
		})
	}
	if err := factory.WriteJSONAtomic(factory.OwnerIndexStaticPath(ownerKey), index); err != nil {
		return err
	}

	composition := ownerFactoryAssetComposition{
		Schema:    "senspace.factory.owner-composition.v1",
		OwnerKey:  ownerKey,
		UpdatedAt: now,
		Relations: make([]ownerFactoryAssetCompositionEdge, 0, len(relations)),
	}
	for _, relation := range relations {
		composition.Relations = append(composition.Relations, ownerFactoryAssetCompositionEdge{
			Id:            strconv.FormatInt(relation.Id, 10),
			RelationType:  relation.RelationType,
			SourceAssetId: strconv.FormatInt(relation.SourceAssetId, 10),
			TargetAssetId: strconv.FormatInt(relation.TargetAssetId, 10),
			Metadata:      decodeRelationMetadata(relation.MetadataJson),
		})
	}
	return factory.WriteJSONAtomic(factory.OwnerCompositionStaticPath(ownerKey), composition)
}

// 写入单个资产静态快照。
func writeFactoryAssetSnapshot(asset factory.Asset) error {
	snapshot := mintedFactoryAsset{
		Schema:       "senspace.factory.asset.v2",
		AssetId:      strconv.FormatInt(asset.Id, 10),
		PluginId:     asset.PluginId,
		ReleaseId:    strconv.FormatInt(asset.ReleaseId, 10),
		Version:      asset.Version,
		RuntimeKind:  defaultRuntimeKind(asset.RuntimeKind),
		AssetKind:    asset.AssetKind,
		TemplateRef:  asset.TemplateRef,
		TemplateId:   asset.TemplateId,
		ReleaseUrl:   releaseStaticURLFromAsset(asset),
		MintRecordId: strconv.FormatInt(asset.MintRecordId, 10),
		MintedAt:     asset.CreatedAt.Format(time.RFC3339Nano),
		TemplateRefs: map[string]string{
			"assetMeta":        "asset.meta.json",
			"defaultWaterMeta": "defaultWaterMeta.json",
		},
	}
	return factory.WriteJSONAtomic(factory.AssetStaticPath(asset.PluginId, asset.Id), snapshot)
}

// 返回资产对应的发布快照地址。
func releaseStaticURLFromAsset(asset factory.Asset) string {
	return factory.FactoryStaticURL("releases", asset.PluginId, fmt.Sprintf("%s-%d", asset.Version, asset.ReleaseId), "release.json")
}

// 解码关系扩展数据。
func decodeRelationMetadata(raw string) any {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" || trimmed == "{}" {
		return nil
	}
	var value any
	if err := json.Unmarshal([]byte(trimmed), &value); err != nil {
		return nil
	}
	return value
}

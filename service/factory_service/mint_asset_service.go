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

// 一组可按等级计价的模板资产。
type assetValueCollection struct {
	Label string `json:"label"`
	// 模板数据位置，例如 defaultWaterMeta.json#fish。
	Ref string `json:"ref"`
	// 模板项中用于区分等级的字段名。
	TierField string `json:"tierField"`
	// 单个等级的选择上限。
	MaxTierCount    int64             `json:"maxTierCount"`
	UnitPriceByTier map[string]string `json:"unitPriceByTier"`
}

// 汇总后端计价和 metaPatch 写入所需数据。
type mintSelectionResult struct {
	RefField     string
	SelectedIds  []string
	ExpectedPaid string
}

// 写入静态目录的个人 NFT 元数据。
type mintedFactoryAsset struct {
	Schema      string                     `json:"schema"`
	AssetId     string                     `json:"assetId"`
	PluginId    string                     `json:"pluginId"`
	ReleaseId   string                     `json:"releaseId"`
	Version     string                     `json:"version"`
	RuntimeKind factory.ReleaseRuntimeKind `json:"runtimeKind"`
	OwnerKey    string                     `json:"ownerKey"`
	ReleaseUrl  string                     `json:"releaseUrl"`
	// 按模板字段记录本次抽取的条目 ID。
	MetaPatch    map[string][]string `json:"metaPatch"`
	PricePaid    string              `json:"pricePaid"`
	MintedAt     string              `json:"mintedAt"`
	TemplateRefs map[string]string   `json:"templateRefs,omitempty"`
}

// 钱包地址对应的静态资产索引。
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
	AssetUrl    string                     `json:"assetUrl"`
	ReleaseUrl  string                     `json:"releaseUrl"`
	MintedAt    string                     `json:"mintedAt"`
}

// 按发布快照中的价值模板生成个人 NFT 元数据。
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
		if release.MintedCount+1 > release.TotalSupply {
			return newConflictError("铸造数量超过可发行数量")
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
		if mintSelection.ExpectedPaid != req.TotalPaid {
			return newParameterError(fmt.Sprintf("支付总额应为 %s", mintSelection.ExpectedPaid))
		}

		record := factory.MintRecord{
			Id:            generateID(),
			PluginId:      release.PluginId,
			ReleaseId:     release.Id,
			UserId:        user.Id,
			WalletAddress: walletAddress,
			Quantity:      1,
			TotalPaid:     req.TotalPaid,
			ChainId:       req.ChainId,
			TxHash:        strings.TrimSpace(req.TxHash),
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		release.MintedCount++
		if release.MintedCount >= release.TotalSupply {
			release.Status = factory.ReleaseStatusSoldOut
		}
		if err := tx.Save(&release).Error; err != nil {
			return err
		}

		if err := ensurePluginOwnership(tx, user.Id, release); err != nil {
			return err
		}

		ownerKey := factory.OwnerIndexKey(walletAddress)
		mintedAt := time.Now().Format(time.RFC3339Nano)
		assetId := record.Id
		assetUrl := factory.AssetStaticURL(release.PluginId, assetId)
		releaseUrl := factory.ReleaseStaticURL(release)
		asset := mintedFactoryAsset{
			Schema:      "senspace.factory.asset.v1",
			AssetId:     strconv.FormatInt(assetId, 10),
			PluginId:    release.PluginId,
			ReleaseId:   strconv.FormatInt(release.Id, 10),
			Version:     release.Version,
			RuntimeKind: defaultRuntimeKind(release.RuntimeKind),
			OwnerKey:    ownerKey,
			ReleaseUrl:  releaseUrl,
			MetaPatch: map[string][]string{
				mintSelection.RefField: mintSelection.SelectedIds,
			},
			PricePaid: req.TotalPaid,
			MintedAt:  mintedAt,
			TemplateRefs: map[string]string{
				"assetMeta":        "asset.meta.json",
				"defaultWaterMeta": "defaultWaterMeta.json",
			},
		}
		if err := factory.WriteJSONAtomic(factory.AssetStaticPath(release.PluginId, assetId), asset); err != nil {
			return err
		}
		if err := appendOwnerFactoryAsset(ownerKey, ownerFactoryAssetEntry{
			AssetId:     strconv.FormatInt(assetId, 10),
			PluginId:    release.PluginId,
			ReleaseId:   strconv.FormatInt(release.Id, 10),
			Version:     release.Version,
			RuntimeKind: defaultRuntimeKind(release.RuntimeKind),
			AssetUrl:    assetUrl,
			ReleaseUrl:  releaseUrl,
			MintedAt:    mintedAt,
		}); err != nil {
			return err
		}

		response = MintAssetResponse{
			AssetId:       strconv.FormatInt(assetId, 10),
			AssetUrl:      assetUrl,
			OwnerIndexUrl: factory.OwnerIndexStaticURL(ownerKey),
			TotalPaid:     req.TotalPaid,
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &response, nil
}

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

func loadReleaseMintTemplate(release factory.Release) (assetValueTemplate, error) {
	dir := factory.ReleaseStaticDir(release)
	var valueTemplate assetValueTemplate
	if err := readJSONFile(filepath.Join(dir, "asset.meta.json"), &valueTemplate); err != nil {
		return valueTemplate, err
	}
	return valueTemplate, nil
}

func readJSONFile(path string, target any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

func resolveMintSelection(
	release factory.Release,
	valueTemplate assetValueTemplate,
	inputs map[string]map[string]int64,
) (mintSelectionResult, error) {
	// 前端以 collection.ref 作为输入 key，后端据此找到对应集合配置。
	collection, counts, ok := selectMintCollection(valueTemplate, inputs)
	if !ok {
		return mintSelectionResult{}, newParameterError("发布模板缺少可铸造资产配置")
	}
	if len(counts) == 0 {
		return mintSelectionResult{}, newParameterError("请选择" + collectionDisplayName(collection) + "的等级和数量")
	}
	templateFile, refField, err := parseAssetRef(collection.Ref)
	if err != nil {
		return mintSelectionResult{}, err
	}
	// # 后面的模板字段会同步写入 metaPatch。
	templateItems, err := loadTemplateItems(filepath.Join(factory.ReleaseStaticDir(release), templateFile), refField)
	if err != nil {
		return mintSelectionResult{}, err
	}
	tierField := strings.TrimSpace(collection.TierField)
	if tierField == "" {
		tierField = "tier"
	}

	total := new(big.Rat)
	base, err := parsePriceRat(valueTemplate.BasePrice, "基础价格")
	if err != nil {
		return mintSelectionResult{}, err
	}
	total.Add(total, base)

	var totalCount int64
	var selected []string
	for tier, count := range counts {
		tier = strings.TrimSpace(tier)
		if tier == "" || count <= 0 {
			continue
		}
		unitRaw, ok := collection.UnitPriceByTier[tier]
		if !ok {
			return mintSelectionResult{}, newParameterError("未知等级: " + tier)
		}
		if collection.MaxTierCount > 0 && count > collection.MaxTierCount {
			return mintSelectionResult{}, newParameterError(fmt.Sprintf("%s等级 %s 不能超过 %d", collectionDisplayName(collection), tier, collection.MaxTierCount))
		}
		unit, err := parsePriceRat(unitRaw, tier+" 单价")
		if err != nil {
			return mintSelectionResult{}, err
		}
		total.Add(total, new(big.Rat).Mul(unit, big.NewRat(count, 1)))
		// 同一等级可重复抽取模板项，形成独立的个人资产实例。
		ids, err := pickTemplateIds(templateItems, tierField, tier, count)
		if err != nil {
			return mintSelectionResult{}, err
		}
		selected = append(selected, ids...)
		totalCount += count
	}
	if totalCount <= 0 {
		return mintSelectionResult{}, newParameterError("请选择" + collectionDisplayName(collection) + "的等级和数量")
	}
	return mintSelectionResult{
		RefField:     refField,
		SelectedIds:  selected,
		ExpectedPaid: normalizeDecimal(total.FloatString(18)),
	}, nil
}

func selectMintCollection(valueTemplate assetValueTemplate, inputs map[string]map[string]int64) (assetValueCollection, map[string]int64, bool) {
	for _, collection := range valueTemplate.Collections {
		ref := strings.TrimSpace(collection.Ref)
		if ref == "" {
			continue
		}
		if counts, ok := inputs[ref]; ok {
			return collection, counts, true
		}
	}
	for _, collection := range valueTemplate.Collections {
		if len(collection.UnitPriceByTier) > 0 {
			return collection, inputs[strings.TrimSpace(collection.Ref)], true
		}
	}
	for _, collection := range valueTemplate.Collections {
		return collection, inputs[strings.TrimSpace(collection.Ref)], true
	}
	return assetValueCollection{}, nil, false
}

func collectionDisplayName(collection assetValueCollection) string {
	label := strings.TrimSpace(collection.Label)
	if label != "" {
		return label
	}
	return strings.TrimSpace(collection.Ref)
}

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

func pickTemplateIds(templates []map[string]any, tierField string, tier string, count int64) ([]string, error) {
	candidates := make([]string, 0, len(templates))
	for _, item := range templates {
		id, _ := item["id"].(string)
		if strings.TrimSpace(id) != "" && fmt.Sprint(item[tierField]) == tier {
			candidates = append(candidates, id)
		}
	}
	if len(candidates) == 0 {
		return nil, newParameterError("等级 " + tier + " 没有可用资产模板")
	}

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

func appendOwnerFactoryAsset(ownerKey string, entry ownerFactoryAssetEntry) error {
	path := factory.OwnerIndexStaticPath(ownerKey)
	index := ownerFactoryAssetIndex{
		Schema:   "senspace.factory.owner-assets.v1",
		OwnerKey: ownerKey,
		Assets:   []ownerFactoryAssetEntry{},
	}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &index); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	for _, existing := range index.Assets {
		if existing.AssetId == entry.AssetId {
			return nil
		}
	}
	index.Schema = "senspace.factory.owner-assets.v1"
	index.OwnerKey = ownerKey
	index.UpdatedAt = time.Now().Format(time.RFC3339Nano)
	index.Assets = append([]ownerFactoryAssetEntry{entry}, index.Assets...)
	return factory.WriteJSONAtomic(path, index)
}

package ds_service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/disintegration/imaging"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"senspace/domain"
	"senspace/domain/ds"
	"senspace/domain/factory"
	"senspace/pkg/app/security"
	"senspace/pkg/setting"
	"senspace/pkg/util"
)

const (
	pluginAssetSchema = "senspace.plugin-assets.v1"
	pluginStateSchema = "senspace.plugin-state.v1"
	defaultImageKind  = "image"
	defaultImageMime  = "image/jpeg"
)

// 资源绑定保存参数。
type PluginAssetBindingInput struct {
	AssetId       uint64          `json:"assetId,string"`   // 资源ID。
	CollectionKey string          `json:"collectionKey"`    // 资源集合键。
	SortOrder     int             `json:"sortOrder"`        // 展示排序。
	Config        json.RawMessage `json:"config,omitempty"` // 资源展示配置。
}

// 插件实例状态保存参数。
type PluginInstanceStateInput struct {
	ExpectedRevision *int64                    `json:"expectedRevision,omitempty"` // 期望状态版本。
	State            map[string]interface{}    `json:"state,omitempty"`            // 插件整体状态。
	Pose             map[string]interface{}    `json:"pose,omitempty"`             // 预留位姿状态。
	Bindings         []PluginAssetBindingInput `json:"bindings,omitempty"`         // 资源绑定配置。
}

// 静态快照返回信息。
type PluginAssetSnapshot struct {
	OwnerKey    string `json:"ownerKey"`    // 钱包索引键。
	FactAssetId string `json:"factAssetId"` // 插件资产实例ID。
	ManifestUrl string `json:"manifestUrl"` // 资源清单地址。
	StateUrl    string `json:"stateUrl"`    // 实例状态地址。
	Revision    int64  `json:"revision"`    // 状态版本号。
}

// 上传资源返回信息。
type PluginAssetUploadResult struct {
	Snapshot PluginAssetSnapshot `json:"snapshot"` // 最新静态快照。
	Asset    pluginAssetJSON     `json:"asset"`    // 新增资源信息。
}

// 静态清单中的单个资源。
type pluginAssetJSON struct {
	AssetId   string `json:"assetId"`            // 资源ID。
	Kind      string `json:"kind"`               // 资源类型。
	Mime      string `json:"mime"`               // 媒体类型。
	Url       string `json:"url"`                // 原图地址。
	ThumbUrl  string `json:"thumbUrl,omitempty"` // 缩略图地址。
	Hash      string `json:"hash"`               // 内容哈希。
	Width     int    `json:"width,omitempty"`    // 图片宽度。
	Height    int    `json:"height,omitempty"`   // 图片高度。
	SizeBytes int64  `json:"sizeBytes"`          // 文件大小。
}

// 资源清单静态文件。
type pluginAssetManifestJSON struct {
	Schema      string            `json:"schema"`      // 协议版本。
	OwnerKey    string            `json:"ownerKey"`    // 钱包索引键。
	FactAssetId string            `json:"factAssetId"` // 插件资产实例ID。
	UpdatedAt   string            `json:"updatedAt"`   // 更新时间。
	Assets      []pluginAssetJSON `json:"assets"`      // 资源列表。
}

// 状态快照中的资源绑定。
type pluginAssetBindingJSON struct {
	AssetId string      `json:"assetId"`          // 资源ID。
	Order   int         `json:"order"`            // 展示排序。
	Config  interface{} `json:"config,omitempty"` // 展示配置。
}

// 插件实例状态静态文件。
type pluginInstanceStateJSON struct {
	Schema      string                              `json:"schema"`      // 协议版本。
	OwnerKey    string                              `json:"ownerKey"`    // 钱包索引键。
	FactAssetId string                              `json:"factAssetId"` // 插件资产实例ID。
	Revision    int64                               `json:"revision"`    // 状态版本号。
	UpdatedAt   string                              `json:"updatedAt"`   // 更新时间。
	State       interface{}                         `json:"state"`       // 插件整体状态。
	Collections map[string][]pluginAssetBindingJSON `json:"collections"` // 分组资源配置。
}

// 处理后的上传图片。
type processedPluginImage struct {
	Original []byte // 原图字节。
	Thumb    []byte // 缩略图字节。
	Mime     string // 输出媒体类型。
	Ext      string // 输出文件后缀。
	Hash     string // 内容哈希。
	Width    int    // 图片宽度。
	Height   int    // 图片高度。
}

// UploadPluginAsset 上传并绑定插件实例资源。
func UploadPluginAsset(user security.JwtUser, factAssetId int64, collectionKey string, fileHeader *multipart.FileHeader) (*PluginAssetUploadResult, error) {
	collectionKey = normalizeCollectionKey(collectionKey)
	if factAssetId <= 0 {
		return nil, errors.New("插件资产实例不能为空")
	}
	if fileHeader == nil {
		return nil, errors.New("上传文件不能为空")
	}

	assetOwner, err := resolveOwnedFactoryAsset(domain.Db, user, factAssetId)
	if err != nil {
		return nil, err
	}
	processed, err := processUploadedPluginImage(fileHeader)
	if err != nil {
		return nil, err
	}

	assetID := generatePluginAssetID()
	assetDir := ds.PluginAssetDir(assetOwner.OwnerKey, factAssetId, assetID)
	originalName := "original" + processed.Ext
	thumbName := "thumb" + processed.Ext
	publicURL := ds.PluginAssetStaticURL(assetOwner.OwnerKey, strconv.FormatInt(factAssetId, 10), "assets", strconv.FormatUint(assetID, 10), originalName)
	thumbURL := ds.PluginAssetStaticURL(assetOwner.OwnerKey, strconv.FormatInt(factAssetId, 10), "assets", strconv.FormatUint(assetID, 10), thumbName)
	storagePath := filepath.Join(assetDir, originalName)
	if err := writeUploadedPluginImageFiles(assetDir, originalName, thumbName, processed); err != nil {
		return nil, err
	}

	var createdAsset ds.PluginAsset
	err = domain.Db.Transaction(func(tx *gorm.DB) error {
		createdAsset = ds.PluginAsset{
			Id:           assetID,
			OwnerKey:     assetOwner.OwnerKey,
			OwnerAddress: assetOwner.OwnerAddress,
			FactAssetId:  factAssetId,
			PluginId:     assetOwner.PluginId,
			ReleaseId:    assetOwner.ReleaseId,
			Kind:         defaultImageKind,
			Mime:         processed.Mime,
			Hash:         processed.Hash,
			SizeBytes:    int64(len(processed.Original)),
			Width:        processed.Width,
			Height:       processed.Height,
			StoragePath:  storagePath,
			PublicUrl:    publicURL,
			ThumbUrl:     thumbURL,
			Status:       ds.PluginAssetStatusActive,
		}
		createdAsset.CreatedBy = user.Id
		createdAsset.UpdatedBy = user.Id
		if err := tx.Create(&createdAsset).Error; err != nil {
			return err
		}
		sortOrder, err := nextPluginAssetSortOrder(tx, factAssetId, collectionKey)
		if err != nil {
			return err
		}
		binding := ds.PluginAssetBinding{
			OwnerKey:      assetOwner.OwnerKey,
			FactAssetId:   factAssetId,
			AssetId:       assetID,
			CollectionKey: collectionKey,
			SortOrder:     sortOrder,
			ConfigJson:    "{}",
			Status:        ds.PluginAssetBindingStatusActive,
		}
		binding.CreatedBy = user.Id
		binding.UpdatedBy = user.Id
		if err := tx.Create(&binding).Error; err != nil {
			return err
		}
		return bumpPluginInstanceState(tx, user.Id, assetOwner.OwnerKey, factAssetId)
	})
	if err != nil {
		_ = os.RemoveAll(assetDir)
		return nil, err
	}
	snapshot, err := RebuildPluginAssetSnapshot(assetOwner.OwnerKey, factAssetId)
	if err != nil {
		return nil, err
	}
	assetJSON := buildPluginAssetJSON(createdAsset)
	return &PluginAssetUploadResult{Snapshot: *snapshot, Asset: assetJSON}, nil
}

// SavePluginInstanceState 保存插件实例状态并刷新静态快照。
func SavePluginInstanceState(user security.JwtUser, factAssetId int64, input PluginInstanceStateInput) (*PluginAssetSnapshot, error) {
	if factAssetId <= 0 {
		return nil, errors.New("插件资产实例不能为空")
	}
	assetOwner, err := resolveOwnedFactoryAsset(domain.Db, user, factAssetId)
	if err != nil {
		return nil, err
	}
	err = domain.Db.Transaction(func(tx *gorm.DB) error {
		state, err := lockPluginInstanceState(tx, assetOwner.OwnerKey, factAssetId)
		if err != nil {
			return err
		}
		if input.ExpectedRevision != nil && state.Revision != *input.ExpectedRevision {
			return fmt.Errorf("状态版本已变化，请刷新后重试")
		}
		if input.Bindings != nil {
			if err := replacePluginAssetBindings(tx, user.Id, assetOwner.OwnerKey, factAssetId, input.Bindings); err != nil {
				return err
			}
		}
		state.OwnerKey = assetOwner.OwnerKey
		state.FactAssetId = factAssetId
		if input.State != nil {
			state.StateJson = encodeJSONOrEmpty(input.State)
		}
		if input.Pose != nil {
			state.PoseJson = encodeJSONOrEmpty(input.Pose)
		}
		state.Revision += 1
		state.UpdatedBy = user.Id
		if state.Id == 0 {
			state.CreatedBy = user.Id
			return tx.Create(&state).Error
		}
		return tx.Save(&state).Error
	})
	if err != nil {
		return nil, err
	}
	return RebuildPluginAssetSnapshot(assetOwner.OwnerKey, factAssetId)
}

// DeletePluginAsset 删除插件资源绑定并刷新静态快照。
func DeletePluginAsset(user security.JwtUser, factAssetId int64, assetID uint64) (*PluginAssetSnapshot, error) {
	if factAssetId <= 0 || assetID == 0 {
		return nil, errors.New("插件资产实例或资源不能为空")
	}
	assetOwner, err := resolveOwnedFactoryAsset(domain.Db, user, factAssetId)
	if err != nil {
		return nil, err
	}
	err = domain.Db.Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&ds.PluginAsset{}).
			Where("id = ? AND owner_key = ? AND fact_asset_id = ?", assetID, assetOwner.OwnerKey, factAssetId).
			Updates(map[string]interface{}{
				"status":     ds.PluginAssetStatusDeleted,
				"updated_by": user.Id,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return errors.New("资源不存在或无权删除")
		}
		if err := tx.Model(&ds.PluginAssetBinding{}).
			Where("asset_id = ? AND owner_key = ? AND fact_asset_id = ?", assetID, assetOwner.OwnerKey, factAssetId).
			Updates(map[string]interface{}{
				"status":     ds.PluginAssetBindingStatusDeleted,
				"updated_by": user.Id,
			}).Error; err != nil {
			return err
		}
		return bumpPluginInstanceState(tx, user.Id, assetOwner.OwnerKey, factAssetId)
	})
	if err != nil {
		return nil, err
	}
	return RebuildPluginAssetSnapshot(assetOwner.OwnerKey, factAssetId)
}

// RebuildPluginAssetSnapshot 重建插件实例资源静态快照。
func RebuildPluginAssetSnapshot(ownerKey string, factAssetId int64) (*PluginAssetSnapshot, error) {
	manifest, state, revision, err := buildPluginAssetSnapshotPayload(domain.Db, ownerKey, factAssetId)
	if err != nil {
		return nil, err
	}
	instanceDir := ds.PluginAssetInstanceDir(ownerKey, factAssetId)
	if err := os.MkdirAll(instanceDir, os.ModePerm); err != nil {
		return nil, err
	}
	if err := writeJSONFileAtomic(ds.PluginAssetManifestPath(ownerKey, factAssetId), manifest); err != nil {
		return nil, err
	}
	if err := writeJSONFileAtomic(ds.PluginAssetStatePath(ownerKey, factAssetId), state); err != nil {
		return nil, err
	}
	return &PluginAssetSnapshot{
		OwnerKey:    ownerKey,
		FactAssetId: strconv.FormatInt(factAssetId, 10),
		ManifestUrl: ds.PluginAssetManifestURL(ownerKey, factAssetId),
		StateUrl:    ds.PluginAssetStateURL(ownerKey, factAssetId),
		Revision:    revision,
	}, nil
}

// ResolvePluginAssetOwner 校验并返回当前用户拥有的插件资产实例。
func ResolvePluginAssetOwner(user security.JwtUser, factAssetId int64) (*factory.Asset, error) {
	return resolveOwnedFactoryAsset(domain.Db, user, factAssetId)
}

func resolveOwnedFactoryAsset(db *gorm.DB, user security.JwtUser, factAssetId int64) (*factory.Asset, error) {
	ownerKey := factory.OwnerIndexKey(user.Addr)
	var asset factory.Asset
	err := db.Where("id = ? AND owner_key = ? AND status = ?", factAssetId, ownerKey, factory.AssetStatusActive).First(&asset).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.New("插件资产实例不存在或无权访问")
	}
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func processUploadedPluginImage(fileHeader *multipart.FileHeader) (*processedPluginImage, error) {
	if fileHeader.Size <= 0 {
		return nil, errors.New("上传文件为空")
	}
	if setting.Config.App.ImageMaxSize > 0 && fileHeader.Size > setting.Config.App.ImageMaxSize {
		return nil, errors.New("图片超过最大限制")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, errors.New("上传文件为空")
	}
	ctype := http.DetectContentType(data[:minInt(len(data), 512)])
	if !strings.HasPrefix(ctype, "image/") {
		return nil, errors.New("不是图片文件")
	}
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if cfg.Width > 10000 || cfg.Height > 10000 {
		return nil, errors.New("图片分辨率太大")
	}
	src, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	if src.Bounds().Dx() > 1600 || src.Bounds().Dy() > 1600 {
		src = imaging.Resize(src, 1600, 0, imaging.CatmullRom)
	}
	mime := defaultImageMime
	ext := ".jpg"
	var original bytes.Buffer
	if hasImageAlpha(src) {
		mime = "image/png"
		ext = ".png"
		if err := png.Encode(&original, src); err != nil {
			return nil, err
		}
	} else if err := jpeg.Encode(&original, src, &jpeg.Options{Quality: 82}); err != nil {
		return nil, err
	}
	thumb := imaging.Thumbnail(src, 360, 360, imaging.CatmullRom)
	var thumbBuffer bytes.Buffer
	if mime == "image/png" {
		if err := png.Encode(&thumbBuffer, thumb); err != nil {
			return nil, err
		}
	} else if err := jpeg.Encode(&thumbBuffer, thumb, &jpeg.Options{Quality: 78}); err != nil {
		return nil, err
	}
	sum := sha256.Sum256(original.Bytes())
	return &processedPluginImage{
		Original: original.Bytes(),
		Thumb:    thumbBuffer.Bytes(),
		Mime:     mime,
		Ext:      ext,
		Hash:     "sha256:" + hex.EncodeToString(sum[:]),
		Width:    src.Bounds().Dx(),
		Height:   src.Bounds().Dy(),
	}, nil
}

func writeUploadedPluginImageFiles(assetDir string, originalName string, thumbName string, processed *processedPluginImage) error {
	if err := os.MkdirAll(assetDir, os.ModePerm); err != nil {
		return err
	}
	if err := writeFileAtomic(filepath.Join(assetDir, originalName), processed.Original); err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(assetDir, thumbName), processed.Thumb)
}

func replacePluginAssetBindings(tx *gorm.DB, userID uint64, ownerKey string, factAssetId int64, bindings []PluginAssetBindingInput) error {
	if err := tx.Model(&ds.PluginAssetBinding{}).
		Where("owner_key = ? AND fact_asset_id = ? AND status = ?", ownerKey, factAssetId, ds.PluginAssetBindingStatusActive).
		Updates(map[string]interface{}{
			"status":     ds.PluginAssetBindingStatusDeleted,
			"updated_by": userID,
		}).Error; err != nil {
		return err
	}
	for _, input := range bindings {
		collectionKey := normalizeCollectionKey(input.CollectionKey)
		var count int64
		if err := tx.Model(&ds.PluginAsset{}).
			Where("id = ? AND owner_key = ? AND fact_asset_id = ? AND status = ?", input.AssetId, ownerKey, factAssetId, ds.PluginAssetStatusActive).
			Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("资源 %d 不存在或已删除", input.AssetId)
		}
		binding := ds.PluginAssetBinding{
			OwnerKey:      ownerKey,
			FactAssetId:   factAssetId,
			AssetId:       input.AssetId,
			CollectionKey: collectionKey,
			SortOrder:     input.SortOrder,
			ConfigJson:    normalizeRawJSON(input.Config),
			Status:        ds.PluginAssetBindingStatusActive,
		}
		binding.CreatedBy = userID
		binding.UpdatedBy = userID
		if err := tx.Create(&binding).Error; err != nil {
			return err
		}
	}
	return nil
}

func bumpPluginInstanceState(tx *gorm.DB, userID uint64, ownerKey string, factAssetId int64) error {
	state, err := lockPluginInstanceState(tx, ownerKey, factAssetId)
	if err != nil {
		return err
	}
	state.OwnerKey = ownerKey
	state.FactAssetId = factAssetId
	state.Revision += 1
	state.UpdatedBy = userID
	if state.Id == 0 {
		state.CreatedBy = userID
		state.StateJson = "{}"
		return tx.Create(&state).Error
	}
	return tx.Save(&state).Error
}

func lockPluginInstanceState(tx *gorm.DB, ownerKey string, factAssetId int64) (ds.PluginInstanceState, error) {
	var state ds.PluginInstanceState
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("owner_key = ? AND fact_asset_id = ?", ownerKey, factAssetId).
		First(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ds.PluginInstanceState{
			OwnerKey:    ownerKey,
			FactAssetId: factAssetId,
			StateJson:   "{}",
			PoseJson:    "{}",
			Revision:    0,
		}, nil
	}
	return state, err
}

func nextPluginAssetSortOrder(tx *gorm.DB, factAssetId int64, collectionKey string) (int, error) {
	var maxOrder int
	err := tx.Model(&ds.PluginAssetBinding{}).
		Where("fact_asset_id = ? AND collection_key = ? AND status = ?", factAssetId, collectionKey, ds.PluginAssetBindingStatusActive).
		Select("COALESCE(MAX(sort_order), 0)").
		Scan(&maxOrder).Error
	return maxOrder + 10, err
}

func buildPluginAssetSnapshotPayload(db *gorm.DB, ownerKey string, factAssetId int64) (*pluginAssetManifestJSON, *pluginInstanceStateJSON, int64, error) {
	var assets []ds.PluginAsset
	if err := db.Where("owner_key = ? AND fact_asset_id = ? AND status = ?", ownerKey, factAssetId, ds.PluginAssetStatusActive).
		Order("id ASC").
		Find(&assets).Error; err != nil {
		return nil, nil, 0, err
	}
	var bindings []ds.PluginAssetBinding
	if err := db.Where("owner_key = ? AND fact_asset_id = ? AND status = ?", ownerKey, factAssetId, ds.PluginAssetBindingStatusActive).
		Order("collection_key ASC, sort_order ASC, id ASC").
		Find(&bindings).Error; err != nil {
		return nil, nil, 0, err
	}
	var state ds.PluginInstanceState
	err := db.Where("owner_key = ? AND fact_asset_id = ?", ownerKey, factAssetId).First(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		state = ds.PluginInstanceState{OwnerKey: ownerKey, FactAssetId: factAssetId, StateJson: "{}", Revision: 0}
	} else if err != nil {
		return nil, nil, 0, err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	manifestAssets := make([]pluginAssetJSON, 0, len(assets))
	for _, asset := range assets {
		manifestAssets = append(manifestAssets, buildPluginAssetJSON(asset))
	}
	collections := map[string][]pluginAssetBindingJSON{}
	for _, binding := range bindings {
		config := decodeJSONValue(binding.ConfigJson)
		item := pluginAssetBindingJSON{
			AssetId: strconv.FormatUint(binding.AssetId, 10),
			Order:   binding.SortOrder,
			Config:  config,
		}
		collections[binding.CollectionKey] = append(collections[binding.CollectionKey], item)
	}
	for key := range collections {
		sort.SliceStable(collections[key], func(i, j int) bool {
			return collections[key][i].Order < collections[key][j].Order
		})
	}
	manifest := &pluginAssetManifestJSON{
		Schema:      pluginAssetSchema,
		OwnerKey:    ownerKey,
		FactAssetId: strconv.FormatInt(factAssetId, 10),
		UpdatedAt:   now,
		Assets:      manifestAssets,
	}
	stateJSON := &pluginInstanceStateJSON{
		Schema:      pluginStateSchema,
		OwnerKey:    ownerKey,
		FactAssetId: strconv.FormatInt(factAssetId, 10),
		Revision:    state.Revision,
		UpdatedAt:   now,
		State:       decodeJSONValue(state.StateJson),
		Collections: collections,
	}
	return manifest, stateJSON, state.Revision, nil
}

func buildPluginAssetJSON(asset ds.PluginAsset) pluginAssetJSON {
	return pluginAssetJSON{
		AssetId:   strconv.FormatUint(asset.Id, 10),
		Kind:      asset.Kind,
		Mime:      asset.Mime,
		Url:       asset.PublicUrl,
		ThumbUrl:  asset.ThumbUrl,
		Hash:      asset.Hash,
		Width:     asset.Width,
		Height:    asset.Height,
		SizeBytes: asset.SizeBytes,
	}
}

func writeJSONFileAtomic(path string, payload interface{}) error {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return writeFileAtomic(path, data)
}

func writeFileAtomic(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpName, path)
}

func normalizeCollectionKey(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "default"
	}
	return trimmed
}

func encodeJSONOrEmpty(value map[string]interface{}) string {
	if value == nil {
		return "{}"
	}
	data, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func normalizeRawJSON(value json.RawMessage) string {
	if len(value) == 0 {
		return "{}"
	}
	var decoded interface{}
	if err := json.Unmarshal(value, &decoded); err != nil {
		return "{}"
	}
	data, err := json.Marshal(decoded)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func decodeJSONValue(raw string) interface{} {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return map[string]interface{}{}
	}
	var decoded interface{}
	if err := json.Unmarshal([]byte(trimmed), &decoded); err != nil {
		return map[string]interface{}{}
	}
	if decoded == nil {
		return map[string]interface{}{}
	}
	return decoded
}

func hasImageAlpha(img image.Image) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			_, _, _, alpha := img.At(x, y).RGBA()
			if alpha != uint32(color.Opaque.A) {
				return true
			}
		}
	}
	return false
}

func generatePluginAssetID() uint64 {
	if util.IdWorker == nil {
		util.Setup()
	}
	return uint64(util.IdWorker.Generate().Int64())
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

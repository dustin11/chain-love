package terrain

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"senspace/domain"
	planet_surface "senspace/domain/planet/surface"
	terrain_domain "senspace/domain/planet/terrain"
	"senspace/domain/sys"
	"senspace/pkg/app/security"
	"senspace/pkg/bizerr"
	road_service "senspace/service/planet/road"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// 当前支持的地形状态结构版本。
	currentSchemaVersion = 4
	// 单个地形状态允许的最大字节数。
	maxStateBytes = 2 * 1024 * 1024
	// 单个星球允许的平台记录上限。
	maxPlatforms = 128
	// 单个星球允许的物件记录上限。
	maxObjects = 5000
)

// 空地形状态的规范 JSON 表示。
var emptyState = json.RawMessage(`{"platforms":[],"objects":[]}`)

const (
	// 地形记录 ID 的最大长度。
	maxRecordIdLength = 128
	// 位置和旋转分量的绝对值上限。
	maxTransformComponent = 1_000_000
	// 缩放分量的最大值。
	maxScaleComponent = 1_000
	// 物件变体种子的最大值。
	maxTerrainVariantSeed = 2_147_483_647
)

// 允许发布的地形物件预设。
var terrainObjectPresetIds = map[string]struct{}{
	"cypress":     {},
	"shrub":       {},
	"grass-clump": {},
	"daisy-patch": {},
	"tulip-patch": {},
	"fern":        {},
	"rock":        {},
	"rock-pillar": {},
	"rock-slab":   {},
	"rock-ridge":  {},
	"pebble-rock": {},
	"box":         {},
	"sphere":      {},
	"cylinder":    {},
	"cone":        {},
}

// 支持单独设置纹理的基础形状预设。
var terrainTextureableObjectPresetIds = map[string]struct{}{
	"box":         {},
	"sphere":      {},
	"cylinder":    {},
	"cone":        {},
	"rock":        {},
	"rock-pillar": {},
	"rock-slab":   {},
	"rock-ridge":  {},
	"pebble-rock": {},
}

// 允许发布的围栏模型。
var fenceModelIds = map[string]struct{}{
	"brick-curb":        {},
	"road-curb":         {},
	"highway-guardrail": {},
	"arrow-barrier":     {},
	"park-railing":      {},
	"brick-wall":        {},
}

// 允许发布的围栏纹理。
var fenceMaterialIds = map[string]struct{}{
	"brick": {},
	"wood":  {},
	"steel": {},
}

// 服务端校验后的独立位置、旋转与缩放。
type terrainTransform struct {
	Position []float64 `json:"position"`
	Rotation []float64 `json:"rotation"`
	Scale    []float64 `json:"scale"`
}

// terrainFence 保存平台外边界围栏样式。
type terrainFence struct {
	ModelId    string `json:"modelId"`
	MaterialId string `json:"materialId"`
}

// 允许发布的平台记录。
type terrainPlatform struct {
	Id          string             `json:"id"`
	Kind        string             `json:"kind"`
	MaterialId  string             `json:"materialId"`
	Transform   terrainTransform   `json:"transform"`
	HeightField terrainHeightField `json:"heightField"`
	Fence       *terrainFence      `json:"fence,omitempty"`
}

// 允许发布的实例物件记录。
type terrainObject struct {
	Id          string           `json:"id"`
	Kind        string           `json:"kind"`
	PlatformId  string           `json:"platformId,omitempty"`
	PresetId    string           `json:"presetId"`
	MaterialId  string           `json:"materialId,omitempty"`
	Transform   terrainTransform `json:"transform"`
	VariantSeed int64            `json:"variantSeed"`
}

// 高度场中的稀疏压缩块。
type terrainHeightChunk struct {
	X        int    `json:"x"`
	Z        int    `json:"z"`
	Encoding string `json:"encoding"`
	Code     *int   `json:"code,omitempty"`
	Data     string `json:"data,omitempty"`
}

// 第一版固定参数的高度场。
type terrainHeightField struct {
	Version         int                  `json:"version"`
	Enabled         bool                 `json:"enabled"`
	BaseHeight      float64              `json:"baseHeight"`
	CellSize        float64              `json:"cellSize"`
	SamplesPerChunk int                  `json:"samplesPerChunk"`
	HeightUnit      float64              `json:"heightUnit"`
	ZeroCode        int                  `json:"zeroCode"`
	Chunks          []terrainHeightChunk `json:"chunks"`
}

// schemaVersion=2 的完整发布载荷。
type terrainState struct {
	Platforms []terrainPlatform `json:"platforms"`
	Objects   []terrainObject   `json:"objects"`
}

// 公开读取一个星球当前发布的地形。
func GetPublished(planetId int) (*DocumentResponse, error) {
	if planetId <= 0 {
		return nil, bizerr.Parameter("planetId无效")
	}
	if domain.Db == nil {
		return nil, fmt.Errorf("planet terrain db not initialized")
	}
	var document terrain_domain.Document
	err := domain.Db.First(&document, "planet_id = ?", planetId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		hash := sha256.Sum256(emptyState)
		return &DocumentResponse{
			SchemaVersion: currentSchemaVersion,
			Revision:      0,
			UpdatedAt:     time.Unix(0, 0).UTC().Format(time.RFC3339Nano),
			ContentHash:   hex.EncodeToString(hash[:]),
			State:         append(json.RawMessage(nil), emptyState...),
		}, nil
	}
	if err != nil {
		return nil, err
	}
	response := mapDocument(document)
	if response.SchemaVersion != document.SchemaVersion ||
		string(response.State) != document.StateJson {
		if err := domain.Db.Model(&terrain_domain.Document{}).
			Where("planet_id = ?", planetId).
			UpdateColumns(map[string]interface{}{
				"schema_version": response.SchemaVersion,
				"state_json":     string(response.State),
				"content_hash":   response.ContentHash,
			}).Error; err != nil {
			return nil, fmt.Errorf("clean automatic terrain data: %w", err)
		}
	}
	return response, nil
}

// 校验星球真实归属并以乐观锁保存地形。
func SavePublished(
	planetId int,
	request SaveRequest,
	user *security.JwtUser,
) (*DocumentResponse, error) {
	if planetId <= 0 {
		return nil, bizerr.Parameter("planetId无效")
	}
	if user == nil || user.Id == 0 {
		return nil, bizerr.Unauthorized()
	}
	if request.ExpectedRevision < 0 {
		return nil, bizerr.Parameter("expectedRevision无效")
	}
	if request.SchemaVersion != currentSchemaVersion {
		return nil, bizerr.Parameter("不支持的地形结构版本")
	}
	normalizedState, err := validateState(request.State)
	if err != nil {
		return nil, err
	}
	if domain.Db == nil {
		return nil, fmt.Errorf("planet terrain db not initialized")
	}

	var saved terrain_domain.Document
	err = domain.Db.Transaction(func(tx *gorm.DB) error {
		var databaseUser sys.User
		if err := tx.Select("id", "planet_id").First(&databaseUser, user.Id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return bizerr.Unauthorized()
			}
			return err
		}
		// 权限以数据库中的 sys_user.planet_id 为准，不能只信任 JWT 快照。
		if databaseUser.PlanetId != planetId {
			return bizerr.Forbidden("只有星球主人可以发布地形")
		}

		var current terrain_domain.Document
		findErr := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&current, "planet_id = ?", planetId).Error
		if errors.Is(findErr, gorm.ErrRecordNotFound) {
			if request.ExpectedRevision != 0 {
				return bizerr.Conflict("地形版本已更新，请重新加载")
			}
			now := time.Now().UTC()
			hash := sha256.Sum256(normalizedState)
			saved = terrain_domain.Document{
				PlanetId:      planetId,
				SchemaVersion: request.SchemaVersion,
				Revision:      1,
				StateJson:     string(normalizedState),
				ContentHash:   hex.EncodeToString(hash[:]),
				CreatInfo: domain.CreatInfo{
					CreatedAt: now,
					CreatedBy: user.Id,
				},
				UpdateInfo: domain.UpdateInfo{
					UpdatedAt: now,
					UpdatedBy: user.Id,
				},
			}
			return tx.Create(&saved).Error
		}
		if findErr != nil {
			return findErr
		}
		if current.Revision != request.ExpectedRevision {
			return bizerr.Conflict("地形版本已更新，请重新加载")
		}
		removedPlatformIds, err := findRemovedPlatformIds(
			json.RawMessage(current.StateJson),
			normalizedState,
		)
		if err != nil {
			return fmt.Errorf("resolve removed terrain platforms: %w", err)
		}

		hash := sha256.Sum256(normalizedState)
		current.SchemaVersion = request.SchemaVersion
		current.Revision++
		current.StateJson = string(normalizedState)
		current.ContentHash = hex.EncodeToString(hash[:])
		current.UpdatedAt = time.Now().UTC()
		current.UpdatedBy = user.Id
		if err := tx.Save(&current).Error; err != nil {
			return err
		}
		if err := road_service.RemovePlatformAnchorsTx(
			tx,
			planetId,
			removedPlatformIds,
			user.Id,
		); err != nil {
			return err
		}
		saved = current
		return nil
	})
	if err != nil {
		return nil, err
	}
	return mapDocument(saved), nil
}

// findRemovedPlatformIds 比较前后状态，只返回本次真正删除的平台。
func findRemovedPlatformIds(
	currentState json.RawMessage,
	nextState json.RawMessage,
) (map[string]struct{}, error) {
	var current terrainState
	if err := json.Unmarshal(currentState, &current); err != nil {
		return nil, err
	}
	var next terrainState
	if err := json.Unmarshal(nextState, &next); err != nil {
		return nil, err
	}
	nextIds := make(map[string]struct{}, len(next.Platforms))
	for _, platform := range next.Platforms {
		nextIds[platform.Id] = struct{}{}
	}
	removedIds := make(map[string]struct{})
	for _, platform := range current.Platforms {
		if _, exists := nextIds[platform.Id]; !exists {
			removedIds[platform.Id] = struct{}{}
		}
	}
	return removedIds, nil
}

// 校验完整结构、有限姿态、预设白名单与记录上限。
func validateState(state json.RawMessage) (json.RawMessage, error) {
	if len(state) == 0 || len(state) > maxStateBytes || !json.Valid(state) {
		return nil, bizerr.Parameter("地形状态无效或过大")
	}
	var envelope terrainState
	decoder := json.NewDecoder(bytes.NewReader(state))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&envelope); err != nil {
		return nil, bizerr.Parameter("地形状态结构无效")
	}
	if err := ensureJsonEnd(decoder); err != nil {
		return nil, bizerr.Parameter("地形状态必须是对象")
	}
	if envelope.Platforms == nil || envelope.Objects == nil {
		return nil, bizerr.Parameter("地形状态必须包含platforms和objects")
	}
	if len(envelope.Platforms) > maxPlatforms {
		return nil, bizerr.Parameter("地形平面数量超过上限")
	}
	if len(envelope.Objects) > maxObjects {
		return nil, bizerr.Parameter("地形物件数量超过上限")
	}

	usedIds := make(map[string]struct{}, len(envelope.Platforms)+len(envelope.Objects))
	platformIds := make(map[string]struct{}, len(envelope.Platforms))
	for _, platform := range envelope.Platforms {
		if platform.Kind != "platform" {
			return nil, bizerr.Parameter("地形平台kind无效")
		}
		if !planet_surface.IsMaterialID(platform.MaterialId) {
			return nil, bizerr.Parameter("地形平台材质无效")
		}
		if err := validateTerrainRecordId(platform.Id, usedIds); err != nil {
			return nil, err
		}
		if !validTerrainTransform(platform.Transform) {
			return nil, bizerr.Parameter("地形平台变换无效")
		}
		if err := validateTerrainHeightField(platform.HeightField); err != nil {
			return nil, err
		}
		if err := validateFence(platform.Fence); err != nil {
			return nil, err
		}
		platformIds[platform.Id] = struct{}{}
	}
	for _, object := range envelope.Objects {
		if object.Kind != "object" {
			return nil, bizerr.Parameter("地形物件kind无效")
		}
		if _, exists := terrainObjectPresetIds[object.PresetId]; !exists {
			return nil, bizerr.Parameter("地形物件预设无效")
		}
		if object.MaterialId != "" {
			if _, exists := terrainTextureableObjectPresetIds[object.PresetId]; !exists {
				return nil, bizerr.Parameter("只有基本形状可以设置纹理")
			}
			if !planet_surface.IsMaterialID(object.MaterialId) {
				return nil, bizerr.Parameter("地形物件纹理无效")
			}
		}
		if object.VariantSeed < 0 || object.VariantSeed > maxTerrainVariantSeed {
			return nil, bizerr.Parameter("地形物件variantSeed无效")
		}
		if object.PlatformId != "" {
			if _, exists := platformIds[object.PlatformId]; !exists {
				return nil, bizerr.Parameter("地形物件所属平台无效")
			}
		}
		if err := validateTerrainRecordId(object.Id, usedIds); err != nil {
			return nil, err
		}
		if !validTerrainTransform(object.Transform) {
			return nil, bizerr.Parameter("地形物件变换无效")
		}
	}
	normalized, err := json.Marshal(envelope)
	if err != nil {
		return nil, fmt.Errorf("marshal terrain state: %w", err)
	}
	return normalized, nil
}

// validateFence 校验可选围栏模型和纹理白名单。
func validateFence(fence *terrainFence) error {
	if fence == nil {
		return nil
	}
	if _, exists := fenceModelIds[fence.ModelId]; !exists {
		return bizerr.Parameter("地形围栏模型无效")
	}
	if _, exists := fenceMaterialIds[fence.MaterialId]; !exists {
		return bizerr.Parameter("地形围栏纹理无效")
	}
	return nil
}

// 校验固定高度场参数、块唯一性与压缩流完整性。
func validateTerrainHeightField(field terrainHeightField) error {
	if field.Version != 2 ||
		field.CellSize != 0.5 ||
		field.SamplesPerChunk != 64 ||
		field.HeightUnit != 0.005 ||
		field.ZeroCode != 32768 ||
		math.IsNaN(field.BaseHeight) ||
		math.IsInf(field.BaseHeight, 0) ||
		math.Abs(field.BaseHeight) > maxTransformComponent ||
		len(field.Chunks) > 4096 {
		return bizerr.Parameter("地形高度场参数无效")
	}
	if field.Enabled != (len(field.Chunks) > 0) {
		return bizerr.Parameter("地形高度场启用状态无效")
	}
	usedKeys := make(map[string]struct{}, len(field.Chunks))
	for _, chunk := range field.Chunks {
		if chunk.X < -32768 || chunk.X > 32768 ||
			chunk.Z < -32768 || chunk.Z > 32768 {
			return bizerr.Parameter("地形高度块坐标无效")
		}
		key := fmt.Sprintf("%d:%d", chunk.X, chunk.Z)
		if _, exists := usedKeys[key]; exists {
			return bizerr.Parameter("地形高度块坐标重复")
		}
		usedKeys[key] = struct{}{}
		switch chunk.Encoding {
		case "constant":
			if chunk.Code == nil || *chunk.Code < 0 || *chunk.Code > 65535 || chunk.Data != "" {
				return bizerr.Parameter("地形常量高度块无效")
			}
			if *chunk.Code == field.ZeroCode {
				return bizerr.Parameter("地形高度块不能是空白基准块")
			}
		case "delta-rle-v1":
			if chunk.Code != nil || len(chunk.Data) == 0 || len(chunk.Data) > 65536 {
				return bizerr.Parameter("地形压缩高度块无效")
			}
			hasHeightChange, err := validateTerrainHeightChunkData(
				chunk.Data,
				field.SamplesPerChunk*field.SamplesPerChunk,
				field.ZeroCode,
			)
			if err != nil {
				return err
			}
			if !hasHeightChange {
				return bizerr.Parameter("地形高度块不能是空白基准块")
			}
		default:
			return bizerr.Parameter("地形高度块编码无效")
		}
	}
	return nil
}

// 解码差分游程并确认高度、样本数和尾部字节均有效。
func validateTerrainHeightChunkData(
	data string,
	expectedSamples int,
	zeroCode int,
) (bool, error) {
	bytes, err := base64.StdEncoding.Strict().DecodeString(data)
	if err != nil {
		return false, bizerr.Parameter("地形高度块Base64无效")
	}
	cursor := 0
	samples := 0
	previous := 32768
	hasHeightChange := false
	for cursor < len(bytes) && samples < expectedSamples {
		runLength, nextCursor, ok := readTerrainVarUint(bytes, cursor)
		if !ok || runLength == 0 || runLength > uint64(expectedSamples-samples) {
			return false, bizerr.Parameter("地形高度块游程无效")
		}
		cursor = nextCursor
		encodedDelta, nextCursor, ok := readTerrainVarUint(bytes, cursor)
		if !ok {
			return false, bizerr.Parameter("地形高度块残差无效")
		}
		cursor = nextCursor
		delta := int64(encodedDelta >> 1)
		if encodedDelta&1 == 1 {
			delta = -delta - 1
		}
		for index := uint64(0); index < runLength; index++ {
			height := int64(previous) + delta
			if height < 0 || height > 65535 {
				return false, bizerr.Parameter("地形高度块采样越界")
			}
			previous = int(height)
			hasHeightChange = hasHeightChange || previous != zeroCode
			samples++
		}
	}
	if samples != expectedSamples || cursor != len(bytes) {
		return false, bizerr.Parameter("地形高度块采样数量无效")
	}
	return hasHeightChange, nil
}

// 从受限字节流读取一个最多五字节的无符号变长整数。
func readTerrainVarUint(data []byte, cursor int) (uint64, int, bool) {
	var result uint64
	for shift := 0; shift <= 28 && cursor < len(data); shift += 7 {
		value := data[cursor]
		cursor++
		result |= uint64(value&0x7f) << shift
		if value&0x80 == 0 {
			return result, cursor, true
		}
	}
	return 0, cursor, false
}

// 确认根对象后不存在第二段 JSON。
func ensureJsonEnd(decoder *json.Decoder) error {
	var extra interface{}
	err := decoder.Decode(&extra)
	if errors.Is(err, io.EOF) {
		return nil
	}
	if err == nil {
		return errors.New("terrain state contains multiple JSON values")
	}
	return err
}

// 校验跨平台和物件唯一的稳定记录 ID。
func validateTerrainRecordId(id string, usedIds map[string]struct{}) error {
	if id == "" || strings.TrimSpace(id) != id || len(id) > maxRecordIdLength {
		return bizerr.Parameter("地形记录id无效")
	}
	if _, exists := usedIds[id]; exists {
		return bizerr.Parameter("地形记录id重复")
	}
	usedIds[id] = struct{}{}
	return nil
}

// 校验位姿数值有限，且缩放分量均为正数。
func validTerrainTransform(transform terrainTransform) bool {
	if !validTerrainVector(transform.Position, false, maxTransformComponent) ||
		!validTerrainVector(transform.Rotation, false, maxTransformComponent) ||
		!validTerrainVector(transform.Scale, true, maxScaleComponent) {
		return false
	}
	return true
}

// 校验精确三分量向量。
func validTerrainVector(values []float64, positive bool, maximum float64) bool {
	if len(values) != 3 {
		return false
	}
	for _, value := range values {
		if math.IsNaN(value) || math.IsInf(value, 0) || math.Abs(value) > maximum {
			return false
		}
		if positive && value <= 0 {
			return false
		}
	}
	return true
}

// 把领域模型映射为稳定 API 信封。
func mapDocument(document terrain_domain.Document) *DocumentResponse {
	updatedAt := document.UpdatedAt.UTC()
	if updatedAt.IsZero() {
		updatedAt = document.CreatedAt.UTC()
	}
	state := json.RawMessage(document.StateJson)
	schemaVersion := document.SchemaVersion
	if schemaVersion < currentSchemaVersion {
		var legacy struct {
			Platforms json.RawMessage `json:"platforms"`
			Objects   json.RawMessage `json:"objects"`
		}
		if json.Unmarshal(state, &legacy) == nil {
			var platforms []map[string]interface{}
			if schemaVersion < 3 && json.Unmarshal(legacy.Platforms, &platforms) == nil {
				for _, platform := range platforms {
					platform["heightField"] = map[string]interface{}{
						"version":         2,
						"enabled":         false,
						"baseHeight":      0.006,
						"cellSize":        0.5,
						"samplesPerChunk": 64,
						"heightUnit":      0.005,
						"zeroCode":        32768,
						"chunks":          []interface{}{},
					}
				}
				legacy.Platforms, _ = json.Marshal(platforms)
			}
			migrated, err := json.Marshal(map[string]json.RawMessage{
				"platforms": legacy.Platforms,
				"objects":   legacy.Objects,
			})
			if err == nil {
				state = migrated
				schemaVersion = currentSchemaVersion
				hash := sha256.Sum256(state)
				document.ContentHash = hex.EncodeToString(hash[:])
			}
		}
	}
	return &DocumentResponse{
		SchemaVersion: schemaVersion,
		Revision:      document.Revision,
		UpdatedAt:     updatedAt.Format(time.RFC3339Nano),
		ContentHash:   document.ContentHash,
		State:         state,
	}
}

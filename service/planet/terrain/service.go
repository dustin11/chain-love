package terrain

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"time"

	"senspace/domain"
	terrain_domain "senspace/domain/planet/terrain"
	"senspace/domain/sys"
	"senspace/pkg/app/security"
	"senspace/pkg/bizerr"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	// 当前支持的地形状态结构版本。
	currentSchemaVersion = 1
	// 单个地形状态允许的最大字节数。
	maxStateBytes        = 2 * 1024 * 1024
	// 单个星球允许的平台记录上限。
	maxPlatforms         = 128
	// 单个星球允许的物件记录上限。
	maxObjects           = 5000
)

// 空地形状态的规范 JSON 表示。
var emptyState = json.RawMessage(`{"platforms":[],"objects":[]}`)

const (
	// 地形记录 ID 的最大长度。
	maxRecordIdLength     = 128
	// 位置和旋转分量的绝对值上限。
	maxTransformComponent = 1_000_000
	// 缩放分量的最大值。
	maxScaleComponent     = 1_000
	// 物件变体种子的最大值。
	maxTerrainVariantSeed = 2_147_483_647
)

// 允许发布的地表材质。
var terrainSurfaceMaterialIds = map[string]struct{}{
	"grass":   {},
	"pebble":  {},
	"marble":  {},
	"asphalt": {},
}

// 允许发布的地形物件预设。
var terrainObjectPresetIds = map[string]struct{}{
	"cypress":     {},
	"shrub":       {},
	"grass-clump": {},
	"daisy-patch": {},
	"tulip-patch": {},
	"fern":        {},
	"rock":        {},
	"pebble-rock": {},
	"box":         {},
	"sphere":      {},
	"cylinder":    {},
	"cone":        {},
}

// 支持单独设置纹理的基础形状预设。
var terrainTextureableObjectPresetIds = map[string]struct{}{
	"box":      {},
	"sphere":   {},
	"cylinder": {},
	"cone":     {},
}

// 服务端校验后的独立位置、旋转与缩放。
type terrainTransform struct {
	Position []float64 `json:"position"`
	Rotation []float64 `json:"rotation"`
	Scale    []float64 `json:"scale"`
}

// 允许发布的平台记录。
type terrainPlatform struct {
	Id         string           `json:"id"`
	Kind       string           `json:"kind"`
	MaterialId string           `json:"materialId"`
	Transform  terrainTransform `json:"transform"`
}

// 允许发布的实例物件记录。
type terrainObject struct {
	Id          string           `json:"id"`
	Kind        string           `json:"kind"`
	PresetId    string           `json:"presetId"`
	MaterialId  string           `json:"materialId,omitempty"`
	Transform   terrainTransform `json:"transform"`
	VariantSeed int64            `json:"variantSeed"`
}

// schemaVersion=1 的完整发布载荷。
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
	return mapDocument(document), nil
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
		saved = current
		return nil
	})
	if err != nil {
		return nil, err
	}
	return mapDocument(saved), nil
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
	for _, platform := range envelope.Platforms {
		if platform.Kind != "platform" {
			return nil, bizerr.Parameter("地形平台kind无效")
		}
		if _, exists := terrainSurfaceMaterialIds[platform.MaterialId]; !exists {
			return nil, bizerr.Parameter("地形平台材质无效")
		}
		if err := validateTerrainRecordId(platform.Id, usedIds); err != nil {
			return nil, err
		}
		if !validTerrainTransform(platform.Transform) {
			return nil, bizerr.Parameter("地形平台变换无效")
		}
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
			if _, exists := terrainSurfaceMaterialIds[object.MaterialId]; !exists {
				return nil, bizerr.Parameter("地形物件纹理无效")
			}
		}
		if object.VariantSeed < 0 || object.VariantSeed > maxTerrainVariantSeed {
			return nil, bizerr.Parameter("地形物件variantSeed无效")
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
	return &DocumentResponse{
		SchemaVersion: document.SchemaVersion,
		Revision:      document.Revision,
		UpdatedAt:     updatedAt.Format(time.RFC3339Nano),
		ContentHash:   document.ContentHash,
		State:         json.RawMessage(document.StateJson),
	}
}

package plugin_share_service

import (
	"compress/gzip"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"senspace/domain"
	"senspace/domain/ds"
	"senspace/domain/sys"
	"senspace/pkg/app/security"
	"senspace/pkg/bizerr"
	"senspace/service/ds_service"

	"gorm.io/gorm"
)

const (
	shareTokenBytes      = 32
	resourceAliasBytes   = 18
	maxSnapshotJSONBytes = 2 * 1024 * 1024
	maximumExpiryHours   = 24 * 365
)

type storedMomentPlugin struct {
	SourceInstanceID string          `json:"sourceInstanceId"`
	SourceSurfaceID  string          `json:"sourceSurfaceId,omitempty"`
	Bootstrap        PluginBootstrap `json:"bootstrap"`
}

type storedMomentSnapshot struct {
	Plugins []storedMomentPlugin `json:"plugins"`
}

// CreateShare 冻结单插件快照。
func CreateShare(user security.JwtUser, input CreateInput) (*CreateResult, error) {
	if domain.Db == nil {
		return nil, errors.New("plugin share db not initialized")
	}
	plugins := normalizeMomentPlugins(input)
	if len(plugins) == 0 {
		return nil, bizerr.Parameter("瞬间至少需要一个插件")
	}
	if len(plugins) > 256 {
		return nil, bizerr.Parameter("瞬间插件数量超过限制")
	}
	primary := plugins[0]
	input.SourceInstanceId = strings.TrimSpace(primary.SourceInstanceId)
	input.SourceSurfaceId = strings.TrimSpace(primary.SourceSurfaceId)
	input.Scope = primary.Scope
	input.Plugin = primary.Plugin
	input.Carrier = primary.Carrier
	input.State = primary.State
	input.ResourceState = primary.ResourceState
	input.SourceSurfaceId = strings.TrimSpace(input.SourceSurfaceId)
	if input.SourceInstanceId == "" || len(input.SourceInstanceId) > 128 {
		return nil, bizerr.Parameter("源插件实例无效")
	}
	if user.Id == 0 {
		return nil, bizerr.Forbidden("当前用户没有可分享的星球")
	}
	currentPlanetID, err := resolveCurrentPlanetID(user.Id, input.SourcePlanetId)
	if err != nil {
		return nil, err
	}
	if currentPlanetID <= 0 {
		return nil, bizerr.Forbidden("当前用户没有可分享的星球")
	}
	if err := validateMomentInput(&input, plugins); err != nil {
		return nil, bizerr.Parameter(err.Error())
	}

	scope, stateJSON, resourceStateJSON, err := resolveSnapshot(user, input)
	if err != nil {
		return nil, err
	}
	if isDynamicPluginDescriptor(input.Plugin) &&
		(scope == nil || scope.Kind != ds.PluginAssetScopeDev) {
		return nil, bizerr.Parameter("动态分享必须来自开发工作区")
	}
	projectedState, err := projectJSON(json.RawMessage(stateJSON), input.SourceInstanceId, input.SourceSurfaceId)
	if err != nil {
		return nil, err
	}
	projectedPlugin, err := projectPluginDescriptor(
		input.Plugin,
		input.SourceInstanceId,
		input.SourceSurfaceId,
	)
	if err != nil {
		return nil, err
	}
	projectedCarrier, err := projectJSON(input.Carrier, input.SourceInstanceId, input.SourceSurfaceId)
	if err != nil {
		return nil, err
	}
	projectedCamera, err := projectJSON(input.Camera, input.SourceInstanceId, input.SourceSurfaceId)
	if err != nil {
		return nil, err
	}
	projectedResourceStateInput, err := projectJSON(
		json.RawMessage(resourceStateJSON),
		input.SourceInstanceId,
		input.SourceSurfaceId,
	)
	if err != nil {
		return nil, err
	}

	token, err := generateOpaqueToken(shareTokenBytes)
	if err != nil {
		return nil, err
	}
	tokenCiphertext, err := encryptManagementToken(token)
	if err != nil {
		return nil, err
	}
	resourceManifest, resourceMap, projectedResourceState, err := buildResourceProjection(
		scope,
		projectedResourceStateInput,
		token,
		input.SourceInstanceId,
		input.SourceSurfaceId,
	)
	if err != nil {
		return nil, err
	}
	storedPlugins := []storedMomentPlugin{{
		SourceInstanceID: input.SourceInstanceId,
		SourceSurfaceID:  input.SourceSurfaceId,
		Bootstrap: PluginBootstrap{
			PluginInstanceId: "moment-plugin-1", SurfaceId: "moment-surface-1",
			PlayerId:         resolvePublicPlayerID(projectedCarrier),
			Plugin:           json.RawMessage(realiasProjection(projectedPlugin, 1)),
			Carrier:          json.RawMessage(realiasProjection(projectedCarrier, 1)),
			State:            json.RawMessage(realiasProjection(projectedState, 1)),
			ResourceState:    json.RawMessage(realiasProjection(projectedResourceState, 1)),
			ResourceManifest: json.RawMessage(resourceManifest),
		},
	}}
	for index := 1; index < len(plugins); index++ {
		stored, mapJSON, buildErr := buildStoredMomentPlugin(user, plugins[index], token, index+1)
		if buildErr != nil {
			return nil, buildErr
		}
		storedPlugins = append(storedPlugins, stored)
		resourceMap, buildErr = mergeJSONObject(resourceMap, mapJSON)
		if buildErr != nil {
			return nil, buildErr
		}
	}
	aliases := map[string]string{}
	for index, stored := range storedPlugins {
		aliases[stored.SourceInstanceID] = fmt.Sprintf("moment-plugin-%d", index+1)
		if stored.SourceSurfaceID != "" {
			aliases[stored.SourceSurfaceID] = fmt.Sprintf("moment-surface-%d", index+1)
		}
	}
	for index := range storedPlugins {
		remapMomentPluginBootstrap(&storedPlugins[index].Bootstrap, aliases)
	}
	snapshotData, err := json.Marshal(storedMomentSnapshot{Plugins: storedPlugins})
	if err != nil {
		return nil, err
	}
	momentScope := strings.ToLower(strings.TrimSpace(input.MomentScope))
	if momentScope == "" {
		momentScope = "plugin"
	}
	shared := input.Shared == nil || *input.Shared
	quotedID, quotedText, err := resolveQuotedMoment(input.QuotedMomentId)
	if err != nil {
		return nil, err
	}
	snapshotKey, err := generateOpaqueToken(resourceAliasBytes)
	if err != nil {
		return nil, err
	}
	snapshotHash := fmt.Sprintf("%x", sha256.Sum256(snapshotData))
	if err := writeMomentSnapshot(snapshotKey, snapshotData); err != nil {
		return nil, err
	}
	expiresAt := resolveExpiry(input.ExpiresInHours)
	share := ds.PluginShare{
		TokenHash:              hashToken(token),
		TokenCiphertext:        tokenCiphertext,
		CreatorUserId:          user.Id,
		SourcePlanetId:         currentPlanetID,
		SourcePluginInstanceId: input.SourceInstanceId,
		SourceSurfaceId:        input.SourceSurfaceId,
		StateJson:              projectedState,
		ResourceStateJson:      projectedResourceState,
		ResourceManifestJson:   resourceManifest,
		ResourceMapJson:        resourceMap,
		PluginDescriptorJson:   projectedPlugin,
		CarrierStateJson:       projectedCarrier,
		CameraStateJson:        projectedCamera,
		MomentScope:            momentScope,
		MomentText:             strings.TrimSpace(input.MomentText),
		IsShared:               shared,
		SnapshotKey:            snapshotKey,
		SnapshotHash:           snapshotHash,
		QuotedMomentId:         quotedID,
		QuotedMomentText:       quotedText,
		Status:                 ds.PluginShareStatusActive,
		ExpiresAt:              expiresAt,
		CreatInfo:              domain.CreatInfo{CreatedBy: user.Id},
		UpdateInfo:             domain.UpdateInfo{UpdatedBy: user.Id},
	}
	applyScopeToShare(&share, scope)
	if err := domain.Db.Create(&share).Error; err != nil {
		_ = os.Remove(momentSnapshotPath(snapshotKey))
		return nil, err
	}
	result := &CreateResult{MomentId: strconv.FormatUint(share.Id, 10), ExpiresAt: share.ExpiresAt}
	if shared {
		result.MomentToken = token
		if momentScope == "planet" {
			result.MomentUrl = "/planet/moment/" + token
		} else {
			result.MomentUrl = "/planet/moment/plugin/" + token
		}
	}
	return result, nil
}

func normalizeMomentPlugins(input CreateInput) []PluginSnapshot {
	if len(input.Plugins) > 0 {
		return input.Plugins
	}
	if len(input.Plugin) == 0 {
		return nil
	}
	return []PluginSnapshot{{SourceInstanceId: input.SourceInstanceId, SourceSurfaceId: input.SourceSurfaceId, Scope: input.Scope, Plugin: input.Plugin, Carrier: input.Carrier, State: input.State, ResourceState: input.ResourceState}}
}

func validateMomentInput(input *CreateInput, plugins []PluginSnapshot) error {
	scope := strings.ToLower(strings.TrimSpace(input.MomentScope))
	if scope != "" && scope != "plugin" && scope != "planet" {
		return errors.New("瞬间范围无效")
	}
	if len([]rune(strings.TrimSpace(input.MomentText))) > 200 {
		return errors.New("瞬间文字不能超过200个字符")
	}
	for _, plugin := range plugins {
		candidate := *input
		candidate.SourceInstanceId, candidate.SourceSurfaceId = plugin.SourceInstanceId, plugin.SourceSurfaceId
		candidate.Scope, candidate.Plugin, candidate.Carrier = plugin.Scope, plugin.Plugin, plugin.Carrier
		candidate.State, candidate.ResourceState = plugin.State, plugin.ResourceState
		if err := validateCreateJSON(candidate); err != nil {
			return err
		}
	}
	return nil
}

func realiasProjection(value string, index int) string {
	value = strings.ReplaceAll(value, "shared-plugin-1", fmt.Sprintf("moment-plugin-%d", index))
	return strings.ReplaceAll(value, "shared-surface-1", fmt.Sprintf("moment-surface-%d", index))
}

func mergeJSONObject(left string, right string) (string, error) {
	merged := map[string]resourceMapEntry{}
	if err := json.Unmarshal([]byte(left), &merged); err != nil {
		return "", err
	}
	other := map[string]resourceMapEntry{}
	if err := json.Unmarshal([]byte(right), &other); err != nil {
		return "", err
	}
	for key, value := range other {
		merged[key] = value
	}
	data, err := json.Marshal(merged)
	return string(data), err
}

func buildStoredMomentPlugin(user security.JwtUser, plugin PluginSnapshot, token string, index int) (storedMomentPlugin, string, error) {
	input := CreateInput{SourceInstanceId: strings.TrimSpace(plugin.SourceInstanceId), SourceSurfaceId: strings.TrimSpace(plugin.SourceSurfaceId), Scope: plugin.Scope, Plugin: plugin.Plugin, Carrier: plugin.Carrier, State: plugin.State, ResourceState: plugin.ResourceState}
	scope, stateJSON, resourceStateJSON, err := resolveSnapshot(user, input)
	if err != nil {
		return storedMomentPlugin{}, "", err
	}
	state, err := projectJSON(json.RawMessage(stateJSON), input.SourceInstanceId, input.SourceSurfaceId)
	if err != nil {
		return storedMomentPlugin{}, "", err
	}
	descriptor, err := projectPluginDescriptor(input.Plugin, input.SourceInstanceId, input.SourceSurfaceId)
	if err != nil {
		return storedMomentPlugin{}, "", err
	}
	carrier, err := projectJSON(input.Carrier, input.SourceInstanceId, input.SourceSurfaceId)
	if err != nil {
		return storedMomentPlugin{}, "", err
	}
	resourceInput, err := projectJSON(json.RawMessage(resourceStateJSON), input.SourceInstanceId, input.SourceSurfaceId)
	if err != nil {
		return storedMomentPlugin{}, "", err
	}
	manifest, resourceMap, resourceState, err := buildResourceProjection(scope, resourceInput, token, input.SourceInstanceId, input.SourceSurfaceId)
	if err != nil {
		return storedMomentPlugin{}, "", err
	}
	return storedMomentPlugin{SourceInstanceID: input.SourceInstanceId, SourceSurfaceID: input.SourceSurfaceId, Bootstrap: PluginBootstrap{PluginInstanceId: fmt.Sprintf("moment-plugin-%d", index), SurfaceId: fmt.Sprintf("moment-surface-%d", index), PlayerId: resolvePublicPlayerID(carrier), Plugin: json.RawMessage(realiasProjection(descriptor, index)), Carrier: json.RawMessage(realiasProjection(carrier, index)), State: json.RawMessage(realiasProjection(state, index)), ResourceState: json.RawMessage(realiasProjection(resourceState, index)), ResourceManifest: json.RawMessage(manifest)}}, resourceMap, nil
}

func remapMomentPluginBootstrap(plugin *PluginBootstrap, aliases map[string]string) {
	plugin.Plugin = remapMomentPluginDescriptor(plugin.Plugin, aliases)
	plugin.Carrier = remapMomentJSON(plugin.Carrier, aliases)
	plugin.State = remapMomentJSON(plugin.State, aliases)
	plugin.ResourceState = remapMomentJSON(plugin.ResourceState, aliases)
}

func remapMomentJSON(raw json.RawMessage, aliases map[string]string) json.RawMessage {
	var value any
	if json.Unmarshal(raw, &value) != nil {
		return raw
	}
	var walk func(any) any
	walk = func(current any) any {
		switch typed := current.(type) {
		case string:
			if alias, ok := aliases[typed]; ok {
				return alias
			}
		case []any:
			for index := range typed {
				typed[index] = walk(typed[index])
			}
		case map[string]any:
			for key, item := range typed {
				typed[key] = walk(item)
			}
		}
		return current
	}
	data, err := json.Marshal(walk(value))
	if err != nil {
		return raw
	}
	return data
}

// ListMyShares 查询当前登录用户创建的分享，不返回源星球、实例和资源作用域信息。
func ListMyShares(user security.JwtUser, query ShareListQuery) (*ShareListResult, error) {
	if domain.Db == nil {
		return nil, errors.New("plugin share db not initialized")
	}
	if user.Id == 0 {
		return nil, bizerr.Forbidden("当前用户没有可管理的分享")
	}
	page := query.Page
	if page <= 0 {
		page = 1
	}
	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	status := strings.ToLower(strings.TrimSpace(query.Status))
	if status != "" && status != "active" && status != "expired" {
		return nil, bizerr.Parameter("分享状态无效")
	}
	if err := deleteLegacyRevokedShares(user.Id); err != nil {
		return nil, err
	}

	dbq := domain.Db.Model(&ds.PluginShare{}).
		Where("creator_user_id = ? AND status = ?", user.Id, ds.PluginShareStatusActive)
	now := time.Now()
	switch status {
	case "active":
		dbq = dbq.Where("expires_at IS NULL OR expires_at > ?", now)
	case "expired":
		dbq = dbq.Where("expires_at IS NOT NULL AND expires_at <= ?", now)
	}
	var total int64
	if err := dbq.Count(&total).Error; err != nil {
		return nil, err
	}
	var shares []ds.PluginShare
	if err := dbq.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&shares).Error; err != nil {
		return nil, err
	}
	items := make([]ShareListItem, 0, len(shares))
	for _, share := range shares {
		item := toShareListItem(share, now)
		items = append(items, item)
	}
	return &ShareListResult{Total: total, Page: page, PageSize: pageSize, Items: items}, nil
}

// DeleteShareByID 按创建者管理 ID 物理删除分享及其专属资源。
func DeleteShareByID(idRaw string, user security.JwtUser) error {
	id, err := strconv.ParseUint(strings.TrimSpace(idRaw), 10, 64)
	if err != nil || id == 0 {
		return bizerr.Parameter("分享 ID 无效")
	}
	if domain.Db == nil {
		return errors.New("plugin share db not initialized")
	}
	var share ds.PluginShare
	err = domain.Db.Where("id = ? AND creator_user_id = ?", id, user.Id).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return bizerr.NotFound("分享不存在")
	}
	if err != nil {
		return err
	}
	return deleteShareRecord(&share)
}

func toShareListItem(share ds.PluginShare, now time.Time) ShareListItem {
	status := "active"
	if share.ExpiresAt != nil && !share.ExpiresAt.After(now) {
		status = "expired"
	}
	item := ShareListItem{
		Id:          strconv.FormatUint(share.Id, 10),
		PluginName:  resolvePluginName(share.PluginDescriptorJson),
		MomentScope: normalizedMomentScope(share.MomentScope),
		MomentText:  share.MomentText,
		Shared:      share.IsShared,
		Status:      status,
		CreatedAt:   share.CreatedAt,
		ExpiresAt:   share.ExpiresAt,
	}
	if share.TokenCiphertext == "" {
		return item
	}
	token, err := decryptManagementToken(share.TokenCiphertext)
	if err != nil {
		return item
	}
	if share.IsShared {
		if normalizedMomentScope(share.MomentScope) == "planet" {
			item.ShareUrl = "/planet/moment/" + token
		} else {
			item.ShareUrl = "/planet/moment/plugin/" + token
		}
	}
	item.MomentUrl = item.ShareUrl
	if !share.IsShared {
		item.MomentUrl = "/planet/moment/private/" + item.Id
	}
	return item
}

// GetOwnedBootstrap 允许创建者通过管理 ID 进入私人瞬间。
func GetOwnedBootstrap(idRaw string, user security.JwtUser) (*Bootstrap, error) {
	id, err := strconv.ParseUint(strings.TrimSpace(idRaw), 10, 64)
	if err != nil || id == 0 {
		return nil, bizerr.Parameter("瞬间 ID 无效")
	}
	var share ds.PluginShare
	if err := domain.Db.Where("id = ? AND creator_user_id = ?", id, user.Id).First(&share).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bizerr.NotFound("瞬间不存在")
		}
		return nil, err
	}
	token, err := decryptManagementToken(share.TokenCiphertext)
	if err != nil {
		return nil, err
	}
	bootstrap, err := GetBootstrap(token, &user)
	if err != nil {
		return nil, err
	}
	data, err := readMomentSnapshot(share.SnapshotKey)
	if err != nil {
		return nil, err
	}
	var snapshot storedMomentSnapshot
	if err := json.Unmarshal(data, &snapshot); err != nil {
		return nil, err
	}
	reverseAliases := map[string]string{}
	for _, stored := range snapshot.Plugins {
		reverseAliases[stored.Bootstrap.PluginInstanceId] = stored.SourceInstanceID
		if stored.SourceSurfaceID != "" {
			reverseAliases[stored.Bootstrap.SurfaceId] = stored.SourceSurfaceID
		}
	}
	for index := range bootstrap.Plugins {
		if index >= len(snapshot.Plugins) {
			break
		}
		bootstrap.Plugins[index].OwnerInstanceId = snapshot.Plugins[index].SourceInstanceID
		remapMomentPluginBootstrap(&bootstrap.Plugins[index], reverseAliases)
	}
	if len(bootstrap.Plugins) > 0 {
		primary := bootstrap.Plugins[0]
		bootstrap.OwnerInstanceId = primary.OwnerInstanceId
		bootstrap.State, bootstrap.Carrier = primary.State, primary.Carrier
	}
	return bootstrap, nil
}

func resolvePluginName(raw string) string {
	var descriptor map[string]any
	if err := json.Unmarshal([]byte(raw), &descriptor); err != nil {
		return "插件分享"
	}
	for _, key := range []string{"name", "pluginName", "pluginId", "factoryId"} {
		if value, ok := descriptor[key].(string); ok && strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return "插件分享"
}

// GetBootstrap 返回最小公开投影，并按当前数据库所有权计算权限。
func GetBootstrap(token string, user *security.JwtUser) (*Bootstrap, error) {
	share, err := findActiveShare(token)
	if err != nil {
		return nil, err
	}
	permissions, err := resolvePermissions(share, user)
	if err != nil {
		return nil, err
	}
	if !share.IsShared && (user == nil || user.Id != share.CreatorUserId) {
		return nil, bizerr.NotFound("瞬间不存在或已失效")
	}
	var snapshot storedMomentSnapshot
	if strings.TrimSpace(share.SnapshotKey) != "" {
		data, err := readMomentSnapshot(share.SnapshotKey)
		if err != nil {
			return nil, err
		}
		if fmt.Sprintf("%x", sha256.Sum256(data)) != share.SnapshotHash {
			return nil, errors.New("瞬间快照校验失败")
		}
		if err := json.Unmarshal(data, &snapshot); err != nil {
			return nil, err
		}
	} else if strings.TrimSpace(share.SnapshotJson) != "" {
		if err := json.Unmarshal([]byte(share.SnapshotJson), &snapshot); err != nil {
			return nil, err
		}
	}
	pluginDescriptor, err := restoreLegacyPluginDescriptor(
		share.PluginDescriptorJson,
		share.SourcePluginInstanceId,
	)
	if err != nil {
		return nil, err
	}
	legacyPlugin := PluginBootstrap{
		PluginInstanceId: "shared-plugin-1",
		SurfaceId:        "shared-surface-1",
		PlayerId:         resolvePublicPlayerID(share.CarrierStateJson),
		Plugin:           json.RawMessage(pluginDescriptor),
		Carrier:          json.RawMessage(share.CarrierStateJson),
		State:            json.RawMessage(share.StateJson),
		ResourceState:    json.RawMessage(share.ResourceStateJson),
		ResourceManifest: json.RawMessage(share.ResourceManifestJson),
	}
	plugins := make([]PluginBootstrap, 0, len(snapshot.Plugins))
	for _, stored := range snapshot.Plugins {
		plugin := stored.Bootstrap
		plugin.Plugin = restoreLegacyMomentDynamicIdentity(
			plugin.Plugin,
			plugin.PluginInstanceId,
			stored.SourceInstanceID,
		)
		plugins = append(plugins, plugin)
	}
	if len(plugins) == 0 {
		plugins = append(plugins, legacyPlugin)
	}
	primary := plugins[0]
	var quoted *QuotedMomentSummary
	if share.QuotedMomentId != nil {
		quoted = &QuotedMomentSummary{MomentId: strconv.FormatUint(*share.QuotedMomentId, 10), MomentText: share.QuotedMomentText, Available: true}
		var count int64
		_ = domain.Db.Model(&ds.PluginShare{}).Where("id = ? AND status = ?", *share.QuotedMomentId, ds.PluginShareStatusActive).Count(&count).Error
		quoted.Available = count > 0
	}
	return &Bootstrap{
		Schema: "senspace.planet-moment.v1", MomentId: strconv.FormatUint(share.Id, 10),
		MomentScope: normalizedMomentScope(share.MomentScope), MomentText: share.MomentText,
		MomentCreatedAt: share.CreatedAt, MomentExpiresAt: share.ExpiresAt, Plugins: plugins,
		PluginInstanceId: primary.PluginInstanceId, SurfaceId: primary.SurfaceId, PlayerId: primary.PlayerId,
		Plugin: primary.Plugin, Carrier: primary.Carrier, Camera: json.RawMessage(share.CameraStateJson),
		State: primary.State, ResourceState: primary.ResourceState, ResourceManifest: primary.ResourceManifest,
		Permissions:  permissions,
		QuotedMoment: quoted,
	}, nil
}

func resolveQuotedMoment(raw string) (*uint64, string, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, "", nil
	}
	id, err := strconv.ParseUint(strings.TrimSpace(raw), 10, 64)
	if err != nil || id == 0 {
		return nil, "", bizerr.Parameter("引用瞬间无效")
	}
	var moment ds.PluginShare
	if err := domain.Db.Select("id", "moment_text", "status").First(&moment, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", bizerr.NotFound("引用瞬间不存在")
		}
		return nil, "", err
	}
	return &id, moment.MomentText, nil
}

func normalizedMomentScope(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "planet") {
		return "planet"
	}
	return "plugin"
}

// DeleteShare 按公开令牌物理删除当前用户创建的分享。
func DeleteShare(token string, user security.JwtUser) error {
	if domain.Db == nil {
		return errors.New("plugin share db not initialized")
	}
	var share ds.PluginShare
	err := domain.Db.Where("token_hash = ?", hashToken(strings.TrimSpace(token))).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return bizerr.NotFound("分享不存在")
	}
	if err != nil {
		return err
	}
	if share.CreatorUserId != user.Id {
		return bizerr.Forbidden("")
	}
	return deleteShareRecord(&share)
}

// deleteShareRecord 统一物理删除分享记录。
func deleteShareRecord(share *ds.PluginShare) error {
	if share == nil {
		return errors.New("分享记录不能为空")
	}
	if domain.Db == nil {
		return errors.New("plugin share db not initialized")
	}
	result := domain.Db.Delete(share)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return bizerr.NotFound("分享不存在")
	}
	if strings.TrimSpace(share.SnapshotKey) != "" {
		_ = os.Remove(momentSnapshotPath(share.SnapshotKey))
	}
	return nil
}

func momentSnapshotPath(key string) string {
	return filepath.Join(ds.PlanetMomentsRoot(), filepath.Base(strings.TrimSpace(key))+".json.gz")
}

func writeMomentSnapshot(key string, data []byte) error {
	root := ds.PlanetMomentsRoot()
	if err := os.MkdirAll(root, 0o750); err != nil {
		return err
	}
	path := momentSnapshotPath(key)
	temporary := path + ".tmp"
	file, err := os.OpenFile(temporary, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o640)
	if err != nil {
		return err
	}
	writer, err := gzip.NewWriterLevel(file, gzip.BestSpeed)
	if err == nil {
		_, err = writer.Write(data)
	}
	if closeErr := writer.Close(); err == nil {
		err = closeErr
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		_ = os.Remove(temporary)
		return err
	}
	return os.Rename(temporary, path)
}

func readMomentSnapshot(key string) ([]byte, error) {
	file, err := os.Open(momentSnapshotPath(key))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer reader.Close()
	return io.ReadAll(io.LimitReader(reader, 16*1024*1024))
}

// deleteLegacyRevokedShares 清理旧版本留下的撤销记录，统一到物理删除语义。
func deleteLegacyRevokedShares(userID uint64) error {
	var shares []ds.PluginShare
	if err := domain.Db.Where(
		"creator_user_id = ? AND status = ?",
		userID,
		ds.PluginShareStatusRevoked,
	).Find(&shares).Error; err != nil {
		return err
	}
	for index := range shares {
		if err := deleteShareRecord(&shares[index]); err != nil {
			return err
		}
	}
	return nil
}

// ResolveCommentInstance 把分享作用域实例映射回原评论实例。
func ResolveCommentInstance(token string, publicInstanceID string, user *security.JwtUser) (string, Permissions, error) {
	share, err := findActiveShare(token)
	if err != nil {
		return "", Permissions{}, err
	}
	permissions, err := resolvePermissions(share, user)
	if err != nil {
		return "", Permissions{}, err
	}
	publicInstanceID = strings.TrimSpace(publicInstanceID)
	if share.SnapshotKey != "" {
		data, readErr := readMomentSnapshot(share.SnapshotKey)
		if readErr != nil {
			return "", Permissions{}, readErr
		}
		var snapshot storedMomentSnapshot
		if json.Unmarshal(data, &snapshot) == nil {
			for _, plugin := range snapshot.Plugins {
				if plugin.Bootstrap.PluginInstanceId == publicInstanceID {
					return plugin.SourceInstanceID, permissions, nil
				}
			}
		}
	}
	if publicInstanceID != "shared-plugin-1" {
		return "", Permissions{}, bizerr.NotFound("瞬间内容不存在")
	}
	return share.SourcePluginInstanceId, permissions, nil
}

// ProjectCommentAnchor 将评论锚点中的真实实例 ID 重写为分享别名。
func ProjectCommentAnchor(anchor map[string]any) map[string]any {
	result := make(map[string]any, len(anchor))
	for key, value := range anchor {
		if strings.EqualFold(key, "instanceId") {
			result[key] = "shared-plugin-1"
			continue
		}
		result[key] = value
	}
	return result
}

// ResolveResource 校验分享令牌和随机别名后返回本地资源文件。
func ResolveResource(token string, alias string) (*ResourceTarget, error) {
	share, err := findActiveShare(token)
	if err != nil {
		return nil, err
	}
	var resourceMap map[string]resourceMapEntry
	if err := json.Unmarshal([]byte(share.ResourceMapJson), &resourceMap); err != nil {
		return nil, err
	}
	target, ok := resourceMap[strings.TrimSpace(alias)]
	if !ok {
		return nil, bizerr.NotFound("分享资源不存在")
	}
	cleanPath := filepath.Clean(target.Path)
	root := filepath.Clean(ds.PluginAssetsRoot())
	relative, err := filepath.Rel(root, cleanPath)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return nil, bizerr.NotFound("分享资源不存在")
	}
	info, err := os.Stat(cleanPath)
	if err != nil || info.IsDir() {
		return nil, bizerr.NotFound("分享资源不存在")
	}
	return &ResourceTarget{Path: cleanPath, Mime: target.Mime}, nil
}

func validateCreateJSON(input CreateInput) error {
	for label, raw := range map[string]json.RawMessage{
		"plugin": input.Plugin, "carrier": input.Carrier, "camera": input.Camera,
		"state": input.State, "resourceState": input.ResourceState,
	} {
		if len(raw) > maxSnapshotJSONBytes {
			return fmt.Errorf("%s超过大小限制", label)
		}
		if err := validJSONObject(raw, label); err != nil {
			return err
		}
	}
	if len(input.Plugin) == 0 || len(input.Carrier) == 0 || len(input.Camera) == 0 {
		return errors.New("插件、承载链和相机快照不能为空")
	}
	var descriptor PluginLoadDescriptor
	if err := json.Unmarshal(input.Plugin, &descriptor); err != nil {
		return err
	}
	if descriptor.Kind != "local" && descriptor.Kind != "factory" && descriptor.Kind != "static" && descriptor.Kind != "dynamic" {
		return errors.New("插件加载类型无效")
	}
	if descriptor.Kind != "dynamic" && strings.TrimSpace(descriptor.FactoryID) == "" {
		return errors.New("插件工厂ID不能为空")
	}
	if descriptor.Kind == "dynamic" &&
		(strings.TrimSpace(descriptor.PluginID) == "" ||
			strings.TrimSpace(descriptor.Version) == "") {
		return errors.New("动态插件标识或版本不能为空")
	}
	return nil
}

func isDynamicPluginDescriptor(raw json.RawMessage) bool {
	var descriptor struct {
		Kind string `json:"kind"`
	}
	return json.Unmarshal(raw, &descriptor) == nil && descriptor.Kind == "dynamic"
}

func resolveSnapshot(user security.JwtUser, input CreateInput) (*ds.PluginAssetScope, string, string, error) {
	if input.Scope == nil {
		state := normalizeRawObject(input.State)
		resourceState := normalizeRawObject(input.ResourceState)
		return nil, state, resourceState, nil
	}
	scope, err := resolveOwnedScope(user, *input.Scope)
	if err != nil {
		return nil, "", "", err
	}
	var state ds.PluginInstanceState
	err = applyScopeFilter(domain.Db.Model(&ds.PluginInstanceState{}), *scope).First(&state).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", "", bizerr.Conflict("插件状态尚未持久化，请稍后重试")
	}
	if err != nil {
		return nil, "", "", err
	}
	return scope, normalizeJSONString(state.StateJson), normalizeJSONString(state.ResourceStateJson), nil
}

func resolveOwnedScope(user security.JwtUser, input SourceScopeInput) (*ds.PluginAssetScope, error) {
	switch input.Kind {
	case ds.PluginAssetScopeFact:
		scope, err := ds_service.ResolveFactPluginAssetScope(user, input.FactAssetId)
		return &scope, err
	case ds.PluginAssetScopeDev:
		scope, err := ds_service.ResolveDevPluginAssetScope(user, input.PluginId, input.Version)
		return &scope, err
	case ds.PluginAssetScopeDraft:
		return nil, bizerr.Parameter("草稿插件不能创建公开分享")
	default:
		return nil, bizerr.Parameter("资源空间类型无效")
	}
}

func buildResourceProjection(scope *ds.PluginAssetScope, resourceStateJSON string, token string, sourceInstanceID string, sourceSurfaceID string) (string, string, string, error) {
	manifest := ResourceManifest{Schema: "senspace.planet-moment-resources.v1", Assets: []ResourceManifestItem{}}
	resourceMap := map[string]resourceMapEntry{}
	collections := map[string][]map[string]any{}
	if scope == nil {
		return marshalResourceProjection(manifest, resourceMap, resourceStateJSON, collections, nil)
	}
	var assets []ds.PluginAsset
	if err := applyScopeFilter(domain.Db.Model(&ds.PluginAsset{}), *scope).
		Where("status = ?", ds.PluginAssetStatusActive).Order("id ASC").Find(&assets).Error; err != nil {
		return "", "", "", err
	}
	assetIDMap := make(map[uint64]string, len(assets))
	for index, asset := range assets {
		publicAssetID := fmt.Sprintf("shared-asset-%d", index+1)
		assetIDMap[asset.Id] = publicAssetID
		originalAlias, err := generateOpaqueToken(resourceAliasBytes)
		if err != nil {
			return "", "", "", err
		}
		resourceMap[originalAlias] = resourceMapEntry{Path: asset.StoragePath, Mime: asset.Mime}
		item := ResourceManifestItem{
			AssetId: publicAssetID, Kind: asset.Kind, Mime: asset.Mime,
			Url:  "/api/v1/planet-moments/" + token + "/resources/" + originalAlias,
			Hash: asset.Hash, SizeBytes: asset.SizeBytes, Width: asset.Width, Height: asset.Height,
		}
		if strings.TrimSpace(asset.ThumbUrl) != "" {
			thumbAlias, aliasErr := generateOpaqueToken(resourceAliasBytes)
			if aliasErr != nil {
				return "", "", "", aliasErr
			}
			thumbPath := filepath.Join(filepath.Dir(asset.StoragePath), filepath.Base(asset.ThumbUrl))
			resourceMap[thumbAlias] = resourceMapEntry{Path: thumbPath, Mime: asset.Mime}
			item.ThumbUrl = "/api/v1/planet-moments/" + token + "/resources/" + thumbAlias
		}
		manifest.Assets = append(manifest.Assets, item)
	}
	var bindings []ds.PluginAssetBinding
	if err := applyScopeFilter(domain.Db.Model(&ds.PluginAssetBinding{}), *scope).
		Order("collection_key ASC, sort_order ASC, id ASC").Find(&bindings).Error; err != nil {
		return "", "", "", err
	}
	for _, binding := range bindings {
		publicAssetID, ok := assetIDMap[binding.AssetId]
		if !ok {
			continue
		}
		var config any
		if json.Valid([]byte(binding.ConfigJson)) {
			_ = json.Unmarshal([]byte(binding.ConfigJson), &config)
			config = projectValue(config, sourceInstanceID, sourceSurfaceID)
		}
		collections[binding.CollectionKey] = append(collections[binding.CollectionKey], map[string]any{
			"assetId": publicAssetID, "order": binding.SortOrder, "config": config,
		})
	}
	return marshalResourceProjection(manifest, resourceMap, resourceStateJSON, collections, assetIDMap)
}

func marshalResourceProjection(manifest ResourceManifest, resourceMap map[string]resourceMapEntry, resourceStateJSON string, collections map[string][]map[string]any, assetIDMap map[uint64]string) (string, string, string, error) {
	manifestData, err := json.Marshal(manifest)
	if err != nil {
		return "", "", "", err
	}
	mapData, err := json.Marshal(resourceMap)
	if err != nil {
		return "", "", "", err
	}
	var state any = map[string]any{}
	if json.Valid([]byte(resourceStateJSON)) {
		_ = json.Unmarshal([]byte(resourceStateJSON), &state)
	}
	state = replaceResourceAssetIDs(state, assetIDMap)
	resourceData, err := json.Marshal(map[string]any{"state": state, "collections": collections})
	if err != nil {
		return "", "", "", err
	}
	return string(manifestData), string(mapData), string(resourceData), nil
}

func replaceResourceAssetIDs(value any, assetIDMap map[uint64]string) any {
	if len(assetIDMap) == 0 {
		return value
	}
	switch typed := value.(type) {
	case string:
		if assetID, ok := parseAssetID(typed); ok {
			if publicID, exists := assetIDMap[assetID]; exists {
				return publicID
			}
		}
		return typed
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, replaceResourceAssetIDs(item, assetIDMap))
		}
		return result
	case map[string]any:
		result := make(map[string]any, len(typed))
		mappedResource := false
		if rawAssetID, ok := typed["assetId"].(string); ok {
			if assetID, valid := parseAssetID(rawAssetID); valid {
				_, mappedResource = assetIDMap[assetID]
			}
		}
		for key, item := range typed {
			// 分享页只能经 shareToken 对应的资源代理读取资源。源页面保存的
			// runtimeUrl/url 可能是临时 blob、鉴权地址或原资源目录，保留它们会让
			// 动态插件优先绕过共享资源地址，导致歌词与封面无法加载。
			if mappedResource && (key == "runtimeUrl" || key == "url") {
				continue
			}
			result[key] = replaceResourceAssetIDs(item, assetIDMap)
		}
		return result
	default:
		return value
	}
}

func findActiveShare(token string) (*ds.PluginShare, error) {
	if domain.Db == nil {
		return nil, errors.New("plugin share db not initialized")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, bizerr.NotFound("分享不存在或已失效")
	}
	var share ds.PluginShare
	err := domain.Db.Where("token_hash = ? AND status = ?", hashToken(token), ds.PluginShareStatusActive).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, bizerr.NotFound("分享不存在或已失效")
	}
	if err != nil {
		return nil, err
	}
	if share.ExpiresAt != nil && !share.ExpiresAt.After(time.Now()) {
		if err := deleteShareRecord(&share); err != nil {
			return nil, err
		}
		return nil, bizerr.NotFound("分享不存在或已失效")
	}
	return &share, nil
}

func resolvePermissions(share *ds.PluginShare, user *security.JwtUser) (Permissions, error) {
	if user == nil || user.Id == 0 {
		return Permissions{}, nil
	}
	var current sys.User
	err := domain.Db.Select("id", "planet_id").First(&current, "id = ?", user.Id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Permissions{}, nil
	}
	if err != nil {
		return Permissions{}, err
	}
	isOwner := isCurrentSharePlanetOwner(current, share)
	return Permissions{IsPlanetOwner: isOwner, CanManageComments: isOwner, CanEditPlugin: false}, nil
}

func resolveCurrentPlanetID(userID uint64, requestedPlanetID int) (int, error) {
	var current sys.User
	err := domain.Db.Select("id", "planet_id").First(&current, "id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, bizerr.Forbidden("当前用户没有可分享的星球")
	}
	if err != nil {
		return 0, err
	}
	return selectShareSourcePlanetID(current.PlanetId, requestedPlanetID)
}

// selectShareSourcePlanetID 优先使用服务端已有星球关系；关系尚未建设时保留当前场景的私有星球标识。
func selectShareSourcePlanetID(storedPlanetID int, requestedPlanetID int) (int, error) {
	if storedPlanetID > 0 {
		if requestedPlanetID > 0 && requestedPlanetID != storedPlanetID {
			return 0, bizerr.Forbidden("当前用户无权分享该星球")
		}
		return storedPlanetID, nil
	}
	if requestedPlanetID <= 0 {
		return 0, bizerr.Forbidden("当前用户没有可分享的星球")
	}
	return requestedPlanetID, nil
}

// isCurrentSharePlanetOwner 在星球关系可用时实时判断所有权，否则仅允许分享创建者管理第一阶段分享。
func isCurrentSharePlanetOwner(current sys.User, share *ds.PluginShare) bool {
	if current.PlanetId > 0 {
		return current.PlanetId == share.SourcePlanetId
	}
	return current.Id > 0 && current.Id == share.CreatorUserId
}

func resolveExpiry(hours int) *time.Time {
	// 0 表示永久；正值才创建可到期瞬间。
	if hours <= 0 {
		return nil
	}
	if hours > maximumExpiryHours {
		hours = maximumExpiryHours
	}
	value := time.Now().Add(time.Duration(hours) * time.Hour)
	return &value
}

func applyScopeToShare(share *ds.PluginShare, scope *ds.PluginAssetScope) {
	if scope == nil {
		return
	}
	share.ScopeKind = scope.Kind
	share.ScopeOwnerKey = scope.OwnerKey
	share.ScopePluginId = scope.PluginId
	share.ScopePluginVersion = scope.PluginVersion
	share.ScopeDraftId = scope.DraftId
	if scope.FactAssetId > 0 {
		value := scope.FactAssetId
		share.ScopeFactAssetId = &value
	}
	if scope.ReleaseId > 0 {
		value := scope.ReleaseId
		share.ScopeReleaseId = &value
	}
}

func applyScopeFilter(query *gorm.DB, scope ds.PluginAssetScope) *gorm.DB {
	switch scope.Kind {
	case ds.PluginAssetScopeFact:
		return query.Where("scope_kind = ? AND fact_asset_id = ?", scope.Kind, scope.FactAssetId)
	case ds.PluginAssetScopeDev:
		return query.Where("scope_kind = ? AND owner_key = ? AND plugin_id = ? AND plugin_version = ?", scope.Kind, scope.OwnerKey, scope.PluginId, scope.PluginVersion)
	case ds.PluginAssetScopeDraft:
		return query.Where("scope_kind = ? AND owner_key = ? AND release_id = ? AND draft_id = ?", scope.Kind, scope.OwnerKey, scope.ReleaseId, scope.DraftId)
	default:
		return query.Where("1 = 0")
	}
}

func normalizeRawObject(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "{}"
	}
	return string(raw)
}

func normalizeJSONString(value string) string {
	value = strings.TrimSpace(value)
	if value == "" || !json.Valid([]byte(value)) {
		return "{}"
	}
	return value
}

func resolvePublicPlayerID(carrierJSON string) string {
	var carrier map[string]any
	if json.Unmarshal([]byte(carrierJSON), &carrier) != nil {
		return ""
	}
	if kind, _ := carrier["surfaceKind"].(string); kind == "sphere" {
		return "shared-player-1"
	}
	if player, ok := carrier["player"].(map[string]any); ok && len(player) > 0 {
		return "shared-player-1"
	}
	return ""
}

func parseAssetID(value string) (uint64, bool) {
	id, err := strconv.ParseUint(strings.TrimSpace(value), 10, 64)
	return id, err == nil && id > 0
}

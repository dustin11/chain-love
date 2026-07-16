package plugin_share_service

import (
	"encoding/json"
	"errors"
	"fmt"
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
	defaultExpiryHours   = 24 * 30
	maximumExpiryHours   = 24 * 365
)

// CreateShare 冻结单插件快照并写入随机背景资源。
func CreateShare(user security.JwtUser, input CreateInput, background []byte, extension string) (*CreateResult, error) {
	if domain.Db == nil {
		return nil, errors.New("plugin share db not initialized")
	}
	input.SourceInstanceId = strings.TrimSpace(input.SourceInstanceId)
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
	if len(background) == 0 {
		return nil, bizerr.Parameter("分享背景不能为空")
	}
	if err := validateCreateJSON(input); err != nil {
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
	backgroundStem, err := generateOpaqueToken(18)
	if err != nil {
		return nil, err
	}
	backgroundKey := backgroundStem + extension
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
		BackgroundKey:          backgroundKey,
		Status:                 ds.PluginShareStatusActive,
		ExpiresAt:              &expiresAt,
		CreatInfo:              domain.CreatInfo{CreatedBy: user.Id},
		UpdateInfo:             domain.UpdateInfo{UpdatedBy: user.Id},
	}
	applyScopeToShare(&share, scope)
	if err := writeBackgroundAtomic(backgroundKey, background); err != nil {
		return nil, err
	}
	if err := domain.Db.Create(&share).Error; err != nil {
		_ = removeBackground(backgroundKey)
		return nil, err
	}
	return &CreateResult{
		ShareToken: token,
		ShareUrl:   "/plugin-share/" + token,
		ExpiresAt:  share.ExpiresAt,
	}, nil
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
		Id:         strconv.FormatUint(share.Id, 10),
		PluginName: resolvePluginName(share.PluginDescriptorJson),
		Status:     status,
		CreatedAt:  share.CreatedAt,
		ExpiresAt:  share.ExpiresAt,
	}
	if share.TokenCiphertext == "" {
		return item
	}
	token, err := decryptManagementToken(share.TokenCiphertext)
	if err != nil {
		return item
	}
	item.ShareUrl = "/plugin-share/" + token
	return item
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
	if err := ensureBackgroundReadable(share.BackgroundKey); err != nil {
		return nil, err
	}
	pluginDescriptor, err := restoreLegacyPluginDescriptor(
		share.PluginDescriptorJson,
		share.SourcePluginInstanceId,
	)
	if err != nil {
		return nil, err
	}
	return &Bootstrap{
		Schema:           "senspace.plugin-share.v1",
		PluginInstanceId: "shared-plugin-1",
		SurfaceId:        "shared-surface-1",
		PlayerId:         resolvePublicPlayerID(share.CarrierStateJson),
		BackgroundUrl:    "/static/plugin-shared/" + share.BackgroundKey,
		Plugin:           json.RawMessage(pluginDescriptor),
		Carrier:          json.RawMessage(share.CarrierStateJson),
		Camera:           json.RawMessage(share.CameraStateJson),
		State:            json.RawMessage(share.StateJson),
		ResourceState:    json.RawMessage(share.ResourceStateJson),
		ResourceManifest: json.RawMessage(share.ResourceManifestJson),
		Permissions:      permissions,
	}, nil
}

// DeleteShare 按公开令牌物理删除当前用户创建的分享及其专属资源。
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

// deleteShareRecord 统一处理分享记录和分享专属静态资源的物理删除。
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
	return removeShareResources(*share)
}

// removeShareResources 只删除分享创建的背景文件，不删除源插件引用的资源文件。
func removeShareResources(share ds.PluginShare) error {
	return removeBackground(share.BackgroundKey)
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
	if strings.TrimSpace(publicInstanceID) != "shared-plugin-1" {
		return "", Permissions{}, bizerr.NotFound("分享内容不存在")
	}
	share, err := findActiveShare(token)
	if err != nil {
		return "", Permissions{}, err
	}
	permissions, err := resolvePermissions(share, user)
	if err != nil {
		return "", Permissions{}, err
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
	manifest := ResourceManifest{Schema: "senspace.plugin-share-resources.v1", Assets: []ResourceManifestItem{}}
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
			Url:  "/plugin-share/" + token + "/resources/" + originalAlias,
			Hash: asset.Hash, SizeBytes: asset.SizeBytes, Width: asset.Width, Height: asset.Height,
		}
		if strings.TrimSpace(asset.ThumbUrl) != "" {
			thumbAlias, aliasErr := generateOpaqueToken(resourceAliasBytes)
			if aliasErr != nil {
				return "", "", "", aliasErr
			}
			thumbPath := filepath.Join(filepath.Dir(asset.StoragePath), filepath.Base(asset.ThumbUrl))
			resourceMap[thumbAlias] = resourceMapEntry{Path: thumbPath, Mime: asset.Mime}
			item.ThumbUrl = "/plugin-share/" + token + "/resources/" + thumbAlias
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

func resolveExpiry(hours int) time.Time {
	if hours <= 0 {
		hours = defaultExpiryHours
	}
	if hours > maximumExpiryHours {
		hours = maximumExpiryHours
	}
	return time.Now().Add(time.Duration(hours) * time.Hour)
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

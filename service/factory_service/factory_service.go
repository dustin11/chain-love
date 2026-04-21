package factory_service

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"senspace/domain/factory"
	factoryvo "senspace/domain/factory/vo"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// 创建发布记录。
func PublishPlugin(authorId uint64, req PublishRequest) (*PublishRecord, error) {
	req, err := validatePublishRequest(req)
	if err != nil {
		return nil, err
	}

	tx, err := db()
	if err != nil {
		return nil, err
	}

	var created factory.Release
	err = tx.Transaction(func(tx *gorm.DB) error {
		// 先校验插件归属，再校验当前仓库最新版本和 manifest 快照。
		if _, err := ensurePluginExists(tx, req.PluginId, authorId); err != nil {
			return err
		}

		repoVersion, versionRoot, err := resolveLatestPluginVersionRoot(req.PluginId)
		if err != nil {
			return err
		}
		if repoVersion != req.Manifest.Version {
			return newConflictError("请求版本不是当前插件仓库的最新版本")
		}

		repoManifest, err := loadManifestFromDir(versionRoot)
		if err != nil {
			return err
		}
		if err := compareManifest(req.Manifest, repoManifest); err != nil {
			return err
		}

		var duplicate int64
		if err := tx.Model(&factory.Release{}).
			Where("plugin_id = ? AND version = ?", req.PluginId, req.Manifest.Version).
			Count(&duplicate).Error; err != nil {
			return err
		}
		if duplicate > 0 {
			return newConflictError("此版本已发布！")
		}

		sourceHash, bundleHash, err := buildHashes(versionRoot, repoManifest)
		if err != nil {
			return err
		}

		now := time.Now()
		// 同一插件只能有一个当前主推版本，发布新版本前先清空旧标记。
		if err := tx.Model(&factory.Release{}).
			Where("plugin_id = ? AND current_release = ?", req.PluginId, true).
			Update("current_release", false).Error; err != nil {
			return err
		}

		created = factory.Release{
			Id:               generateID(),
			PluginId:         req.PluginId,
			AuthorId:         authorId,
			Name:             repoManifest.Name,
			Version:          repoManifest.Version,
			Status:           factory.ReleaseStatusPublished,
			ReviewStatus:     factory.ReviewStatusApproved,
			CurrentRelease:   true,
			ManifestSnapshot: toManifestSnapshot(repoManifest),
			Summary:          req.Release.Summary,
			Category:         req.Release.Category,
			Tags:             factory.StringList(req.Release.Tags),
			CoverUrl:         req.Release.CoverUrl,
			TotalSupply:      req.Release.TotalSupply,
			MintPer:          req.Release.MintPer,
			MintPrice:        req.Release.MintPrice,
			MintedCount:      0,
			SourceHash:       sourceHash,
			BundleHash:       bundleHash,
			UpgradePolicy:    req.Release.UpgradePolicy,
			UpgradePrice:     zeroIfEmpty(req.Release.UpgradePrice),
			PublishedAt:      nowPtr(now),
			UpdatedAt:        now,
		}

		return tx.Create(&created).Error
	})
	if err != nil {
		return nil, err
	}

	record := mapRelease(created)
	return &record, nil
}

// 查询我的发布。
func ListMyReleases(authorId uint64, query ReleaseQuery) ([]PublishRecord, error) {
	tx, err := db()
	if err != nil {
		return nil, err
	}

	dbq := tx.Model(&factory.Release{}).Where("author_id = ?", authorId)
	if query.PluginId != "" {
		dbq = dbq.Where("plugin_id = ?", strings.TrimSpace(query.PluginId))
	}
	if query.Status != "" {
		dbq = dbq.Where("status = ?", strings.TrimSpace(query.Status))
	}
	if query.CurrentOnly != nil {
		dbq = dbq.Where("current_release = ?", *query.CurrentOnly)
	}

	var releases []factory.Release
	if err := dbq.Order("published_at desc").Order("updated_at desc").Find(&releases).Error; err != nil {
		return nil, err
	}

	result := make([]PublishRecord, 0, len(releases))
	for _, item := range releases {
		result = append(result, mapRelease(item))
	}
	return result, nil
}

// 查询市场发布。
func ListMarketReleases(query MarketQuery) ([]PublishRecord, error) {
	tx, err := db()
	if err != nil {
		return nil, err
	}

	currentOnly := true
	if query.CurrentOnly != nil {
		currentOnly = *query.CurrentOnly
	}

	dbq := tx.Model(&factory.Release{})
	if currentOnly {
		dbq = dbq.Where("current_release = ?", true)
	}
	if query.PluginId != "" {
		dbq = dbq.Where("plugin_id = ?", strings.TrimSpace(query.PluginId))
	}
	if query.Category != "" {
		dbq = dbq.Where("category = ?", strings.TrimSpace(query.Category))
	}
	if query.Status != "" {
		dbq = dbq.Where("status = ?", strings.TrimSpace(query.Status))
	} else {
		dbq = dbq.Where("status IN ?", []factory.ReleaseStatus{
			factory.ReleaseStatusPublished,
			factory.ReleaseStatusPaused,
			factory.ReleaseStatusSoldOut,
		})
	}

	var releases []factory.Release
	if err := dbq.Order("published_at desc").Order("updated_at desc").Find(&releases).Error; err != nil {
		return nil, err
	}

	filterTags := normalizeTags(query.Tags)
	result := make([]PublishRecord, 0, len(releases))
	for _, item := range releases {
		if len(filterTags) > 0 && !containsAllTags([]string(item.Tags), filterTags) {
			continue
		}
		result = append(result, mapRelease(item))
	}
	return result, nil
}

// 查询发布详情。
func GetReleaseDetail(id string) (*PublishDetail, error) {
	tx, err := db()
	if err != nil {
		return nil, err
	}

	releaseId, err := parseID(id, "发布记录ID")
	if err != nil {
		return nil, err
	}

	var release factory.Release
	if err := tx.First(&release, "id = ?", releaseId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, newNotFoundError("发布记录不存在")
		}
		return nil, err
	}

	var history []factory.ReleasePriceHistory
	if err := tx.Where("release_id = ?", release.Id).
		Order("changed_at desc").
		Find(&history).Error; err != nil {
		return nil, err
	}

	detail := mapReleaseDetail(release, history)
	return &detail, nil
}

// 更新市场信息。
func UpdateReleaseMarket(authorId uint64, req UpdateReleaseRequest) (*PublishRecord, error) {
	req.Id = strings.TrimSpace(req.Id)
	req.Market = normalizeMarket(req.Market)
	if req.Market.Summary == "" {
		return nil, newParameterError("市场摘要不能为空")
	}
	if req.Market.Category == "" {
		return nil, newParameterError("分类不能为空")
	}

	tx, err := db()
	if err != nil {
		return nil, err
	}

	releaseId, err := parseID(req.Id, "发布记录ID")
	if err != nil {
		return nil, err
	}

	var release factory.Release
	if err := tx.First(&release, "id = ?", releaseId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, newNotFoundError("发布记录不存在")
		}
		return nil, err
	}
	if release.AuthorId != authorId {
		return nil, newForbiddenError("无权修改该发布记录")
	}

	release.Summary = req.Market.Summary
	release.Category = req.Market.Category
	release.Tags = factory.StringList(req.Market.Tags)
	release.CoverUrl = req.Market.CoverUrl
	if err := tx.Save(&release).Error; err != nil {
		return nil, err
	}

	record := mapRelease(release)
	return &record, nil
}

// 更新铸造价格。
func UpdateReleasePrice(authorId uint64, req UpdateReleasePriceRequest) (*PublishRecord, error) {
	req.Id = strings.TrimSpace(req.Id)
	price, err := validateDecimalString(req.MintPrice, "铸造价格", false)
	if err != nil {
		return nil, err
	}
	req.MintPrice = price

	tx, err := db()
	if err != nil {
		return nil, err
	}

	releaseId, err := parseID(req.Id, "发布记录ID")
	if err != nil {
		return nil, err
	}

	var release factory.Release
	err = tx.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&release, "id = ?", releaseId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return newNotFoundError("发布记录不存在")
			}
			return err
		}
		if release.AuthorId != authorId {
			return newForbiddenError("无权修改该发布记录")
		}
		if release.MintPrice == req.MintPrice {
			return nil
		}

		// 调价只影响后续铸造，因此必须保留价格变更历史。
		history := factory.ReleasePriceHistory{
			Id:                generateID(),
			ReleaseId:         release.Id,
			PluginId:          release.PluginId,
			PreviousMintPrice: release.MintPrice,
			NextMintPrice:     req.MintPrice,
			Reason:            strings.TrimSpace(req.Reason),
			ChangedBy:         strconv.FormatUint(authorId, 10),
		}
		release.MintPrice = req.MintPrice
		if err := tx.Save(&release).Error; err != nil {
			return err
		}
		return tx.Create(&history).Error
	})
	if err != nil {
		return nil, err
	}

	record := mapRelease(release)
	return &record, nil
}

// 更新发布状态。
func UpdateReleaseStatus(authorId uint64, req UpdateReleaseStatusRequest) (*PublishRecord, error) {
	req.Id = strings.TrimSpace(req.Id)
	switch req.Status {
	case factory.ReleaseStatusPublished, factory.ReleaseStatusPaused, factory.ReleaseStatusClosed:
	default:
		return nil, newParameterError("目标状态不支持")
	}

	tx, err := db()
	if err != nil {
		return nil, err
	}

	releaseId, err := parseID(req.Id, "发布记录ID")
	if err != nil {
		return nil, err
	}

	var release factory.Release
	err = tx.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&release, "id = ?", releaseId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return newNotFoundError("发布记录不存在")
			}
			return err
		}
		if release.AuthorId != authorId {
			return newForbiddenError("无权修改该发布记录")
		}
		if release.Status == factory.ReleaseStatusClosed && req.Status == factory.ReleaseStatusPublished {
			return newConflictError("已关闭的发布记录不能重新上架")
		}
		if release.Status == req.Status {
			return nil
		}

		// 状态变化需要完整留痕，便于市场展示与后续审计。
		history := factory.ReleaseStatusHistory{
			Id:             generateID(),
			ReleaseId:      release.Id,
			PluginId:       release.PluginId,
			PreviousStatus: release.Status,
			NextStatus:     req.Status,
			Reason:         strings.TrimSpace(req.Reason),
			ChangedBy:      strconv.FormatUint(authorId, 10),
		}

		now := time.Now()
		release.Status = req.Status
		switch req.Status {
		case factory.ReleaseStatusPublished:
			if release.PublishedAt == nil {
				release.PublishedAt = nowPtr(now)
			}
		case factory.ReleaseStatusPaused:
			release.PausedAt = nowPtr(now)
		case factory.ReleaseStatusClosed:
			release.ClosedAt = nowPtr(now)
			if release.CurrentRelease {
				release.CurrentRelease = false
				// 当前主推版本被关闭后，尝试回补一个可展示的候选版本。
				if err := promoteFallbackCurrentRelease(tx, release.PluginId, release.Id); err != nil {
					return err
				}
			}
		}

		if err := tx.Save(&release).Error; err != nil {
			return err
		}
		return tx.Create(&history).Error
	})
	if err != nil {
		return nil, err
	}

	record := mapRelease(release)
	return &record, nil
}

// 记录铸造。
func RecordMint(req RecordMintRequest) (*factory.MintRecord, error) {
	releaseId, err := parseID(req.ReleaseId, "发布记录ID")
	if err != nil {
		return nil, err
	}
	if req.UserId == 0 {
		return nil, newParameterError("用户ID不能为空")
	}
	if strings.TrimSpace(req.WalletAddress) == "" {
		return nil, newParameterError("钱包地址不能为空")
	}
	if req.Quantity <= 0 {
		return nil, newParameterError("铸造数量必须大于0")
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

	var record factory.MintRecord
	err = tx.Transaction(func(tx *gorm.DB) error {
		// 对发布记录加锁，防止并发铸造导致超卖。
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
		if req.Quantity > release.MintPer {
			return newParameterError("铸造数量不能超过单次最大铸造量")
		}
		if release.MintedCount+req.Quantity > release.TotalSupply {
			return newConflictError("铸造数量超过可发行数量")
		}

		record = factory.MintRecord{
			Id:            generateID(),
			PluginId:      release.PluginId,
			ReleaseId:     release.Id,
			UserId:        req.UserId,
			WalletAddress: strings.TrimSpace(req.WalletAddress),
			Quantity:      req.Quantity,
			TotalPaid:     req.TotalPaid,
			ChainId:       req.ChainId,
			TxHash:        strings.TrimSpace(req.TxHash),
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}

		release.MintedCount += req.Quantity
		if release.MintedCount >= release.TotalSupply {
			release.Status = factory.ReleaseStatusSoldOut
		}
		if err := tx.Save(&release).Error; err != nil {
			return err
		}

		// 资产表只保存长期事实，首次铸造后写入 minted/effective 两个发布指针。
		var ownership factory.UserOwnership
		err := tx.Where("user_id = ? AND plugin_id = ?", req.UserId, release.PluginId).First(&ownership).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ownership = factory.UserOwnership{
				Id:                 generateID(),
				UserId:             req.UserId,
				PluginId:           release.PluginId,
				MintedReleaseId:    release.Id,
				EffectiveReleaseId: release.Id,
			}
			return tx.Create(&ownership).Error
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// 查询我的资产。
func ListMyOwnerships(userId uint64) ([]UserPluginOwnershipView, error) {
	tx, err := db()
	if err != nil {
		return nil, err
	}

	var ownerships []factory.UserOwnership
	if err := tx.Where("user_id = ?", userId).Order("created_at desc").Find(&ownerships).Error; err != nil {
		return nil, err
	}

	result := make([]UserPluginOwnershipView, 0, len(ownerships))
	for _, ownership := range ownerships {
		view, err := buildOwnershipView(tx, ownership)
		if err != nil {
			return nil, err
		}
		result = append(result, view)
	}
	return result, nil
}

// 升级资产。
func UpgradeOwnership(userId uint64, req UpgradeOwnershipRequest) (*UpgradeRecord, error) {
	ownershipId, err := parseID(req.Id, "资产记录ID")
	if err != nil {
		return nil, err
	}
	targetReleaseId, err := parseID(req.ToReleaseId, "目标发布记录ID")
	if err != nil {
		return nil, err
	}

	tx, err := db()
	if err != nil {
		return nil, err
	}

	var upgraded factory.UpgradeRecord
	err = tx.Transaction(func(tx *gorm.DB) error {
		// 锁定资产记录，避免并发升级覆盖 effectiveReleaseId。
		var ownership factory.UserOwnership
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&ownership, "id = ?", ownershipId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return newNotFoundError("资产记录不存在")
			}
			return err
		}
		if ownership.UserId != userId {
			return newForbiddenError("无权升级该资产")
		}

		var currentRelease factory.Release
		if err := tx.First(&currentRelease, "id = ?", ownership.EffectiveReleaseId).Error; err != nil {
			return err
		}
		var targetRelease factory.Release
		if err := tx.First(&targetRelease, "id = ?", targetReleaseId).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return newNotFoundError("目标发布记录不存在")
			}
			return err
		}
		if targetRelease.PluginId != ownership.PluginId {
			return newConflictError("目标发布记录与资产插件不匹配")
		}
		if compareVersions(targetRelease.Version, currentRelease.Version) <= 0 {
			return newConflictError("只能升级到更高版本")
		}
		switch targetRelease.Status {
		case factory.ReleaseStatusPublished, factory.ReleaseStatusPaused, factory.ReleaseStatusSoldOut:
		default:
			return newConflictError("目标发布记录当前不可升级")
		}

		upgradeType, paidAmount, allowed := releaseAllowsUpgrade(currentRelease.Version, targetRelease)
		if !allowed {
			return newConflictError("当前版本不允许升级到该目标版本")
		}

		now := time.Now()
		// 升级后只更新 effectiveReleaseId，保留 mintedReleaseId 作为原始铸造事实。
		ownership.EffectiveReleaseId = targetRelease.Id
		ownership.UpgradedAt = nowPtr(now)
		if err := tx.Save(&ownership).Error; err != nil {
			return err
		}

		upgraded = factory.UpgradeRecord{
			Id:            generateID(),
			OwnershipId:   ownership.Id,
			UserId:        ownership.UserId,
			PluginId:      ownership.PluginId,
			FromReleaseId: currentRelease.Id,
			ToReleaseId:   targetRelease.Id,
			UpgradeType:   upgradeType,
			PaidAmount:    paidAmount,
			UpgradedAt:    now,
		}
		return tx.Create(&upgraded).Error
	})
	if err != nil {
		return nil, err
	}

	record := mapUpgradeRecord(upgraded)
	return &record, nil
}

// 组装资产视图。
func buildOwnershipView(tx *gorm.DB, ownership factory.UserOwnership) (factoryvo.UserPluginOwnershipView, error) {
	var mintedRelease factory.Release
	if err := tx.First(&mintedRelease, "id = ?", ownership.MintedReleaseId).Error; err != nil {
		return factoryvo.UserPluginOwnershipView{}, err
	}
	var effectiveRelease factory.Release
	if err := tx.First(&effectiveRelease, "id = ?", ownership.EffectiveReleaseId).Error; err != nil {
		return factoryvo.UserPluginOwnershipView{}, err
	}

	latestRelease, err := findCurrentRelease(tx, ownership.PluginId)
	if err != nil {
		return factoryvo.UserPluginOwnershipView{}, err
	}
	return mapOwnership(ownership, mintedRelease.Version, effectiveRelease.Version, latestRelease), nil
}

// 查询当前主推版本。
func findCurrentRelease(tx *gorm.DB, pluginId string) (*factory.Release, error) {
	var release factory.Release
	err := tx.Where("plugin_id = ? AND current_release = ?", pluginId, true).
		Order("published_at desc").
		First(&release).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &release, nil
}

// 回补主推版本。
func promoteFallbackCurrentRelease(tx *gorm.DB, pluginId string, excludedId int64) error {
	var candidates []factory.Release
	if err := tx.Where("plugin_id = ? AND id <> ? AND status IN ?", pluginId, excludedId, []factory.ReleaseStatus{
		factory.ReleaseStatusPublished,
		factory.ReleaseStatusPaused,
		factory.ReleaseStatusSoldOut,
		factory.ReleaseStatusArchived,
	}).Find(&candidates).Error; err != nil {
		return err
	}
	if len(candidates) == 0 {
		return nil
	}

	best := candidates[0]
	for _, candidate := range candidates[1:] {
		if compareVersions(candidate.Version, best.Version) > 0 {
			best = candidate
		}
	}
	return tx.Model(&factory.Release{}).Where("id = ?", best.Id).Update("current_release", true).Error
}

// 判断标签是否全部命中。
func containsAllTags(haystack []string, needles []string) bool {
	index := make(map[string]struct{}, len(haystack))
	for _, tag := range haystack {
		index[tag] = struct{}{}
	}
	for _, tag := range needles {
		if _, ok := index[tag]; !ok {
			return false
		}
	}
	return true
}

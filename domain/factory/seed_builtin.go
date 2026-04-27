package factory

import (
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
)

const builtinFactoryAuthorId uint64 = 0

type builtinFactoryReleaseSeed struct {
	Id          int64
	PluginId    string
	Name        string
	Version     string
	RuntimeKind ReleaseRuntimeKind
	Description string
	Summary     string
	Category    string
	Tags        StringList
	CoverUrl    string
	TotalSupply int64
	MintPer     int64
	MintPrice   string
}

var builtinFactoryReleaseSeeds = []builtinFactoryReleaseSeed{
	{
		Id:          910000000000001,
		PluginId:    "FishTank",
		Name:        "水空间",
		Version:     "1.0.0",
		RuntimeKind: ReleaseRuntimeKindBuiltin,
		Description: "鱼缸插件。",
		Summary:     "可互动，配合管道，可带你进入鱼儿的水世界",
		Category:    "环境",
		Tags:        StringList{"鱼缸", "水世界"},
		CoverUrl:    "/logo512.webp",
		TotalSupply: 10000,
		MintPer:     1,
		MintPrice:   "10",
	},
	{
		Id:          910000000000002,
		PluginId:    "BookPreview",
		Name:        "Book",
		Version:     "1.0.0",
		RuntimeKind: ReleaseRuntimeKindBook,
		Description: "主项目内置书本预览。",
		Summary:     "铸造属于你的书籍体验",
		Category:    "内容",
		Tags:        StringList{"书本", "阅读"},
		CoverUrl:    "/logo512.webp",
		TotalSupply: 100000,
		MintPer:     10,
		MintPrice:   "0",
	},
}

// SeedBuiltinReleases 将宿主内置插件登记为发布记录。
func SeedBuiltinReleases(db *gorm.DB) {
	if db == nil {
		return
	}

	for _, seed := range builtinFactoryReleaseSeeds {
		if err := seedBuiltinRelease(db, seed); err != nil {
			log.Printf("seed builtin factory release %s@%s failed: %v", seed.PluginId, seed.Version, err)
		}
	}
}

func seedBuiltinRelease(db *gorm.DB, seed builtinFactoryReleaseSeed) error {
	now := time.Now()
	release := Release{
		Id:             seed.Id,
		PluginId:       seed.PluginId,
		AuthorId:       builtinFactoryAuthorId,
		AuthorSnapshot: AuthorSnapshot{Id: "senspace", Name: "Senspace"},
		Name:           seed.Name,
		Version:        seed.Version,
		Status:         ReleaseStatusPublished,
		ReviewStatus:   ReviewStatusApproved,
		CurrentRelease: true,
		ManifestSnapshot: PluginManifestSnapshot{
			Name:        seed.Name,
			Version:     seed.Version,
			Entry:       seed.PluginId,
			Description: seed.Description,
		},
		Summary:       seed.Summary,
		Category:      seed.Category,
		Tags:          seed.Tags,
		CoverUrl:      seed.CoverUrl,
		TotalSupply:   seed.TotalSupply,
		MintPer:       seed.MintPer,
		MintPrice:     seed.MintPrice,
		BuildStatus:   BuildStatusReady,
		RuntimeKind:   seed.RuntimeKind,
		UpgradePolicy: ReleaseUpgradePolicyNone,
		UpgradePrice:  "0",
		PublishedAt:   &now,
		BuiltAt:       &now,
	}

	var existing Release
	err := db.Where("plugin_id = ? AND version = ?", seed.PluginId, seed.Version).First(&existing).Error
	if err == nil {
		return updateBuiltinRelease(db, existing, release)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	if err := db.Create(&release).Error; err != nil {
		return err
	}
	return EnsureReleaseStaticSnapshot(release)
}

func updateBuiltinRelease(db *gorm.DB, existing Release, release Release) error {
	updates := map[string]interface{}{
		"plugin_id":         release.PluginId,
		"author_id":         release.AuthorId,
		"author_snapshot":   release.AuthorSnapshot,
		"name":              release.Name,
		"version":           release.Version,
		"status":            release.Status,
		"review_status":     release.ReviewStatus,
		"current_release":   release.CurrentRelease,
		"manifest_snapshot": release.ManifestSnapshot,
		"summary":           release.Summary,
		"category":          release.Category,
		"tags":              release.Tags,
		"cover_url":         release.CoverUrl,
		"total_supply":      release.TotalSupply,
		"mint_per":          release.MintPer,
		"mint_price":        release.MintPrice,
		"build_status":      release.BuildStatus,
		"runtime_kind":      release.RuntimeKind,
		"upgrade_policy":    release.UpgradePolicy,
		"upgrade_price":     release.UpgradePrice,
	}
	if existing.PublishedAt == nil {
		updates["published_at"] = release.PublishedAt
	}
	if existing.BuiltAt == nil {
		updates["built_at"] = release.BuiltAt
	}
	if err := db.Model(&Release{}).Where("id = ?", existing.Id).Updates(updates).Error; err != nil {
		return err
	}
	release.Id = existing.Id
	return EnsureReleaseStaticSnapshot(release)
}

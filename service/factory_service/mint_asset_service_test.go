package factory_service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"senspace/domain"
	"senspace/domain/d_util"
	"senspace/domain/factory"
	"senspace/pkg/app/security"
	"senspace/pkg/setting"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// 独立 NFT 铸造与组合快照。
func TestMintReleaseAssetCreatesComposableNFTSnapshots(t *testing.T) {
	env := setupMintAssetServiceTest(t)
	user := security.JwtUser{
		Id:   990000000000101,
		Addr: "0xe9a51481d67ca775d19b02a220833ce6c4575f53",
	}
	ownerKey := factory.OwnerIndexKey(user.Addr)
	release := createFishTankMintTestRelease(t, env.db)
	t.Cleanup(func() {
		cleanupMintTestData(t, env.db, release.Id, user.Id, ownerKey)
	})

	response, err := MintReleaseAsset(user, strconv.FormatInt(release.Id, 10), MintAssetRequest{
		Inputs: map[string]map[string]int64{
			"defaultWaterMeta.json#tanks": {
				"default": 1,
			},
			"defaultWaterMeta.json#fish": {
				"common": 2,
				"rare":   1,
			},
		},
		TotalPaid: "60",
	})
	require.NoError(t, err)
	require.Equal(t, "60", response.TotalPaid)
	require.Len(t, response.Assets, 4)
	require.Equal(t, factory.OwnerIndexStaticURL(ownerKey), response.OwnerIndexUrl)
	require.Equal(t, factory.OwnerCompositionStaticURL(ownerKey), response.OwnerCompositionUrl)

	var assets []factory.Asset
	require.NoError(t, env.db.Where("release_id = ?", release.Id).Find(&assets).Error)
	require.Len(t, assets, 4)
	require.Equal(t, int64(1), countAssetsByKind(assets, factory.AssetKindTank))
	require.Equal(t, int64(3), countAssetsByKind(assets, factory.AssetKindFish))

	var record factory.MintRecord
	require.NoError(t, env.db.First(&record, "release_id = ? AND user_id = ?", release.Id, user.Id).Error)
	require.Equal(t, int64(4), record.Quantity)
	require.Equal(t, "60.000000000000000000", record.TotalPaid)

	var refreshedRelease factory.Release
	require.NoError(t, env.db.First(&refreshedRelease, "id = ?", release.Id).Error)
	require.Equal(t, int64(4), refreshedRelease.MintedCount)

	var relations []factory.AssetRelation
	require.NoError(t, env.db.Where("owner_key = ?", ownerKey).Find(&relations).Error)
	require.Len(t, relations, 3)
	for _, relation := range relations {
		require.Equal(t, "contains", relation.RelationType)
		require.NotZero(t, relation.SourceAssetId)
		require.NotZero(t, relation.TargetAssetId)
	}

	var index ownerFactoryAssetIndex
	readJSONFileForTest(t, factory.OwnerIndexStaticPath(ownerKey), &index)
	require.Equal(t, "senspace.factory.owner-assets.v2", index.Schema)
	require.Equal(t, ownerKey, index.OwnerKey)
	require.Len(t, index.Assets, 4)

	var composition ownerFactoryAssetComposition
	readJSONFileForTest(t, factory.OwnerCompositionStaticPath(ownerKey), &composition)
	require.Equal(t, "senspace.factory.owner-composition.v1", composition.Schema)
	require.Equal(t, ownerKey, composition.OwnerKey)
	require.Len(t, composition.Relations, 3)

	for _, asset := range assets {
		var snapshot map[string]any
		readJSONFileForTest(t, factory.AssetStaticPath(asset.PluginId, asset.Id), &snapshot)
		require.Equal(t, "senspace.factory.asset.v2", snapshot["schema"])
		require.Equal(t, string(asset.AssetKind), snapshot["assetKind"])
		require.Equal(t, asset.TemplateId, snapshot["templateId"])
		require.NotContains(t, snapshot, "ownerKey")
	}
}

type mintAssetServiceTestEnv struct {
	db *gorm.DB
}

// 本地 ldev MySQL。
func setupMintAssetServiceTest(t *testing.T) mintAssetServiceTestEnv {
	t.Helper()

	originalConfig := *setting.Config
	originalDB := domain.Db
	originalWd, err := os.Getwd()
	require.NoError(t, err)

	repoRoot := factoryServiceRepoRoot(t)
	require.NoError(t, os.Chdir(repoRoot))

	setting.Config.Database = setting.Database{
		Type:     "mysql",
		User:     envOrDefault("SENSPACE_TEST_DB_USER", "root"),
		Password: envOrDefault("SENSPACE_TEST_DB_PASSWORD", "smart@vserp"),
		Host:     envOrDefault("SENSPACE_TEST_DB_HOST", "127.0.0.1:3307"),
		Name:     envOrDefault("SENSPACE_TEST_DB_NAME", "senspace"),
	}
	setting.Config.App.FilePath.Factory = filepath.Join(t.TempDir(), "factory")
	setting.Config.App.RuntimeRootPath = filepath.Join(t.TempDir(), "runtime")

	require.NoError(t, d_util.EnsureDatabaseExists(setting.Config.Database.Name))
	domain.Setup()
	require.NotNil(t, domain.Db, "无法连接测试数据库，请确认 ldev MySQL 运行在 127.0.0.1:3307")
	d_util.InitTable(domain.Db)

	t.Cleanup(func() {
		if domain.Db != nil {
			if sqlDB, err := domain.Db.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}
		domain.Db = originalDB
		*setting.Config = originalConfig
		require.NoError(t, os.Chdir(originalWd))
	})

	return mintAssetServiceTestEnv{db: domain.Db}
}

// 创建 FishTank 铸造测试发布。
func createFishTankMintTestRelease(t *testing.T, db *gorm.DB) factory.Release {
	t.Helper()

	now := time.Now()
	release := factory.Release{
		Id:             generateID(),
		PluginId:       "FishTank",
		AuthorId:       990000000000101,
		AuthorSnapshot: factory.AuthorSnapshot{Id: "mint-test", Name: "Mint Test"},
		Name:           "FishTank Mint Test",
		Version:        "test-" + strconv.FormatInt(time.Now().UnixNano(), 10),
		Status:         factory.ReleaseStatusPublished,
		ReviewStatus:   factory.ReviewStatusApproved,
		CurrentRelease: false,
		ManifestSnapshot: factory.PluginManifestSnapshot{
			Name:        "FishTank Mint Test",
			Version:     "test",
			Entry:       "FishTank",
			Description: "组合资产测试",
		},
		Summary:       "组合资产测试",
		Category:      "test",
		Tags:          factory.StringList{"test"},
		TotalSupply:   100,
		MintPer:       10,
		MintPrice:     "0",
		BuildStatus:   factory.BuildStatusReady,
		RuntimeKind:   factory.ReleaseRuntimeKindBuiltin,
		UpgradePolicy: factory.ReleaseUpgradePolicyNone,
		UpgradePrice:  "0",
		PublishedAt:   &now,
		BuiltAt:       &now,
	}
	require.NoError(t, db.Create(&release).Error)
	return release
}

// 清理铸造测试数据。
func cleanupMintTestData(t *testing.T, db *gorm.DB, releaseId int64, userId uint64, ownerKey string) {
	t.Helper()

	require.NoError(t, db.Exec("DELETE FROM fact_asset_relation WHERE owner_key = ?", ownerKey).Error)
	require.NoError(t, db.Exec("DELETE FROM fact_asset WHERE release_id = ?", releaseId).Error)
	require.NoError(t, db.Exec("DELETE FROM fact_user_ownership WHERE user_id = ? AND plugin_id = ?", userId, "FishTank").Error)
	require.NoError(t, db.Exec("DELETE FROM fact_mint_record WHERE release_id = ?", releaseId).Error)
	require.NoError(t, db.Exec("DELETE FROM fact_release WHERE id = ?", releaseId).Error)
}

// 统计指定类型资产数量。
func countAssetsByKind(assets []factory.Asset, kind factory.AssetKind) int64 {
	var count int64
	for _, asset := range assets {
		if asset.AssetKind == kind {
			count++
		}
	}
	return count
}

// 读取测试 JSON 文件。
func readJSONFileForTest(t *testing.T, path string, target any) {
	t.Helper()

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, json.Unmarshal(data, target))
}

// 定位 factory service 仓库根目录。
func factoryServiceRepoRoot(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)
	for {
		if _, err := os.Stat(filepath.Join(wd, "go.mod")); err == nil {
			return wd
		}
		next := filepath.Dir(wd)
		require.NotEqual(t, wd, next, "未找到仓库根目录")
		wd = next
	}
}

// 返回环境变量或默认值。
func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

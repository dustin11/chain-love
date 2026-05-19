package factory_service

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"senspace/domain"
	"senspace/domain/d_util"
	"senspace/domain/dev"
	"senspace/domain/factory"
	"senspace/pkg/app/security"
	"senspace/pkg/setting"

	"github.com/stretchr/testify/require"
)

func TestReadRuntimeBuildResultValidatesManifest(t *testing.T) {
	outputDir := t.TempDir()
	runtimeDir := filepath.Join(outputDir, "runtime")
	require.NoError(t, os.MkdirAll(runtimeDir, 0o755))

	entryContent := []byte("export default { pluginId: 'demo-plugin' };\n")
	require.NoError(t, os.WriteFile(filepath.Join(runtimeDir, "sandbox-entry.js"), entryContent, 0o644))
	bundleHash, integrity := runtimeHashes(entryContent)

	req := PluginBuildRequest{
		PluginId:  "demo-plugin",
		Version:   "1.2.3",
		ReleaseId: 10001,
	}

	resultWithoutManifest, err := readRuntimeBuildResult(outputDir, req)
	require.Nil(t, resultWithoutManifest)
	require.ErrorContains(t, err, "读取运行清单失败")

	manifest := runtimeManifestFile{
		PluginId:     req.PluginId,
		Version:      req.Version,
		ReleaseId:    "10001",
		BundleHash:   bundleHash,
		Integrity:    integrity,
		SandboxEntry: "runtime/sandbox-entry.js",
		ExternalDependencies: []runtimeManifestDependency{
			{Name: "three", Mode: "external"},
		},
		BundledDependencies: []runtimeManifestDependency{
			{Name: "lodash-es", Mode: "bundled"},
		},
	}
	manifestBytes, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(outputDir, "runtime-manifest.json"), manifestBytes, 0o644))

	result, err := readRuntimeBuildResult(outputDir, req)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, manifest.BundleHash, result.BundleHash)
	require.Equal(t, manifest.Integrity, result.Integrity)
	require.Equal(t, []string{"three"}, result.ExternalDependencies)
	require.Equal(t, []string{"lodash-es"}, result.BundledDependencies)
}

func TestReadRuntimeBuildResultRejectsInvalidDependencyMode(t *testing.T) {
	outputDir := t.TempDir()
	runtimeDir := filepath.Join(outputDir, "runtime")
	require.NoError(t, os.MkdirAll(runtimeDir, 0o755))

	entryContent := []byte("export default { pluginId: 'demo-plugin' };\n")
	require.NoError(t, os.WriteFile(filepath.Join(runtimeDir, "sandbox-entry.js"), entryContent, 0o644))
	bundleHash, integrity := runtimeHashes(entryContent)

	req := PluginBuildRequest{
		PluginId:  "demo-plugin",
		Version:   "1.2.3",
		ReleaseId: 10001,
	}

	manifest := runtimeManifestFile{
		PluginId:     req.PluginId,
		Version:      req.Version,
		ReleaseId:    "10001",
		BundleHash:   bundleHash,
		Integrity:    integrity,
		SandboxEntry: "runtime/sandbox-entry.js",
		ExternalDependencies: []runtimeManifestDependency{
			{Name: "three", Mode: "bundled"},
		},
	}
	manifestBytes, err := json.Marshal(manifest)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(outputDir, "runtime-manifest.json"), manifestBytes, 0o644))

	result, err := readRuntimeBuildResult(outputDir, req)
	require.Nil(t, result)
	require.ErrorContains(t, err, "模式非法")
}

func TestExecuteReleaseBuildWritesReleaseSnapshotForArtifactPlugin(t *testing.T) {
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
		Name:     envOrDefault("SENSPACE_TEST_DB_NAME", "senspace_factory_test"),
	}
	setting.Config.App.FilePath.Factory = filepath.Join(t.TempDir(), "factory")
	setting.Config.App.RuntimeRootPath = filepath.Join(t.TempDir(), "runtime")
	require.NoError(t, d_util.EnsureDatabaseExists(setting.Config.Database.Name))
	domain.Setup()
	require.NotNil(t, domain.Db)
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

	restoreExecutor := SetPluginBuildExecutorForTest(buildExecutorStub(func(req PluginBuildRequest) (*PluginBuildResult, error) {
		return &PluginBuildResult{
			BundleHash: "sha256:test-bundle",
			Integrity:  "sha384:test-integrity",
			BuiltAt:    time.Now(),
		}, nil
	}))
	defer restoreExecutor()

	release := factory.Release{
		Id:           generateID(),
		PluginId:     "ArtifactOnlyPlugin",
		AuthorId:     990000000000104,
		Name:         "Artifact Only Plugin",
		Version:      "1.0.0-" + time.Now().Format("20060102150405"),
		Status:       factory.ReleaseStatusDraft,
		ReviewStatus: factory.ReviewStatusApproved,
		ManifestSnapshot: factory.PluginManifestSnapshot{
			Name:        "Artifact Only Plugin",
			Version:     "1.0.0",
			Entry:       "src/index.ts",
			Description: "test",
		},
		Summary:       "test",
		Category:      "test",
		TotalSupply:   10,
		MintPer:       1,
		MintPrice:     "0",
		BuildStatus:   factory.BuildStatusPending,
		RuntimeKind:   factory.ReleaseRuntimeKindArtifact,
		UpgradePolicy: factory.ReleaseUpgradePolicyNone,
		UpgradePrice:  "0",
	}
	require.NoError(t, domain.Db.Create(&release).Error)
	t.Cleanup(func() {
		require.NoError(t, domain.Db.Exec("DELETE FROM sys_async_task WHERE biz_type = ? AND biz_id = ?", staticTaskBizTypeFactoryRelease, release.Id).Error)
		require.NoError(t, domain.Db.Exec("DELETE FROM fact_release_status_history WHERE release_id = ?", release.Id).Error)
		require.NoError(t, domain.Db.Exec("DELETE FROM fact_release_price_history WHERE release_id = ?", release.Id).Error)
		require.NoError(t, domain.Db.Exec("DELETE FROM fact_release WHERE id = ?", release.Id).Error)
	})

	builtRelease, err := executeReleaseBuild(release)
	require.NoError(t, err)
	require.Equal(t, factory.BuildStatusReady, builtRelease.BuildStatus)

	releaseManifestPath := filepath.Join(factory.ReleaseStaticDir(builtRelease), "release.json")
	var manifest factory.ReleaseStaticManifest
	readJSONFileForTest(t, releaseManifestPath, &manifest)
	require.Equal(t, builtRelease.PluginId, manifest.PluginId)
	require.Equal(t, builtRelease.Version, manifest.Version)
	require.EqualValues(t, builtRelease.TotalSupply, manifest.TotalSupply)
	require.EqualValues(t, builtRelease.MintPer, manifest.MintPer)
	require.Equal(t, builtRelease.MintPrice, manifest.MintPrice)
	require.Equal(t, "0.000000000000000000", manifest.BasePrice)
}

func TestPublishPluginCleansSnapshotWhenVersionAlreadyPublished(t *testing.T) {
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
		Name:     envOrDefault("SENSPACE_TEST_DB_NAME", "senspace_factory_test"),
	}
	tempRoot := t.TempDir()
	setting.Config.App.FilePath.Factory = filepath.Join(tempRoot, "factory")
	setting.Config.App.RuntimeRootPath = filepath.Join(tempRoot, "runtime")
	setting.Config.App.PluginSourceRoot = filepath.Join(tempRoot, "runtime", "plugin-source")
	setting.Config.App.PluginRuntimeRoot = filepath.Join(tempRoot, "runtime", "plugin-runtime")
	require.NoError(t, d_util.EnsureDatabaseExists(setting.Config.Database.Name))
	domain.Setup()
	require.NotNil(t, domain.Db)
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

	plugin := dev.Plugin{
		Id:          990000000000001123,
		Name:        "Published Collision Plugin",
		Description: "test",
		Version:     "1.0.0",
		Author:      "tester",
	}
	plugin.CreatedBy = 990000000000001123
	require.NoError(t, domain.Db.Create(&plugin).Error)
	t.Cleanup(func() {
		_ = domain.Db.Exec("DELETE FROM dev_plugin WHERE id = ?", plugin.Id).Error
	})

	versionRoot := filepath.Join(setting.Config.App.FilePath.Plugin, plugin.Name, "1.0.0")
	require.NoError(t, os.MkdirAll(filepath.Join(versionRoot, "src"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(versionRoot, "manifest.json"), []byte(`{
  "name": "Published Collision Plugin",
  "version": "1.0.0",
  "entry": "src/index.ts"
}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(versionRoot, "src", "index.ts"), []byte("export default {};\n"), 0o644))

	now := time.Now()
	existing := factory.Release{
		Id:           generateID(),
		PluginId:     plugin.Name,
		AuthorId:     plugin.CreatedBy,
		Name:         "Published Collision Plugin",
		Version:      "1.0.0",
		Status:       factory.ReleaseStatusPublished,
		ReviewStatus: factory.ReviewStatusApproved,
		ManifestSnapshot: factory.PluginManifestSnapshot{
			Name:        "Published Collision Plugin",
			Version:     "1.0.0",
			Entry:       "src/index.ts",
			Description: "test",
		},
		Summary:       "test",
		Category:      "test",
		TotalSupply:   10,
		MintPer:       1,
		MintPrice:     "0",
		BuildStatus:   factory.BuildStatusReady,
		RuntimeKind:   factory.ReleaseRuntimeKindArtifact,
		UpgradePolicy: factory.ReleaseUpgradePolicyNone,
		UpgradePrice:  "0",
		PublishedAt:   &now,
		BuiltAt:       &now,
	}
	require.NoError(t, domain.Db.Create(&existing).Error)
	t.Cleanup(func() {
		_ = domain.Db.Exec("DELETE FROM fact_release_status_history WHERE release_id = ?", existing.Id).Error
		_ = domain.Db.Exec("DELETE FROM fact_release_price_history WHERE release_id = ?", existing.Id).Error
		_ = domain.Db.Exec("DELETE FROM fact_release WHERE id = ?", existing.Id).Error
	})

	_, err = PublishPlugin(security.JwtUser{Id: plugin.CreatedBy}, PublishRequest{
		PluginId: plugin.Name,
		Manifest: PluginManifest{
			Name:    "Published Collision Plugin",
			Version: "1.0.0",
			Entry:   "src/index.ts",
		},
		Release: ReleasePayload{
			MutableMarketMetadata: MutableMarketMetadata{
				Summary:  "test",
				Category: "test",
			},
			TotalSupply: 10,
			MintPer:     1,
			MintPrice:   "0",
		},
	})
	require.ErrorContains(t, err, "此版本已发布")

	sourceRoot := filepath.Join(setting.Config.App.PluginSourceRoot, plugin.Name)
	entries, statErr := os.ReadDir(sourceRoot)
	if os.IsNotExist(statErr) {
		return
	}
	require.NoError(t, statErr)
	require.Len(t, entries, 0)
}

type buildExecutorStub func(req PluginBuildRequest) (*PluginBuildResult, error)

func (stub buildExecutorStub) Build(_ context.Context, req PluginBuildRequest) (*PluginBuildResult, error) {
	return stub(req)
}

func runtimeHashes(entryContent []byte) (string, string) {
	sha256Sum := sha256.Sum256(entryContent)
	sha384Sum := sha512.Sum384(entryContent)
	return "sha256:" + hex.EncodeToString(sha256Sum[:]),
		"sha384-" + base64.StdEncoding.EncodeToString(sha384Sum[:])
}

package factory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"senspace/pkg/setting"

	"github.com/stretchr/testify/require"
)

func TestActivateReleaseStaticSnapshotCanRollbackAndCommit(t *testing.T) {
	oldFactoryRoot := setting.Config.App.FilePath.Factory
	oldRuntimeRoot := setting.Config.App.RuntimeRootPath
	setting.Config.App.FilePath.Factory = t.TempDir()
	setting.Config.App.RuntimeRootPath = ""
	t.Cleanup(func() {
		setting.Config.App.FilePath.Factory = oldFactoryRoot
		setting.Config.App.RuntimeRootPath = oldRuntimeRoot
	})

	release := Release{
		Id:       123,
		PluginId: "TestPlugin",
		Version:  "1.0.0",
	}
	finalDir := ReleaseStaticDir(release)
	stagingDir := ReleaseStaticStagingDir(release)
	require.NoError(t, os.MkdirAll(finalDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(finalDir, "old.json"), []byte("{}\n"), 0644))
	require.NoError(t, os.MkdirAll(stagingDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(stagingDir, "new.json"), []byte("{}\n"), 0644))

	backupDir, err := ActivateReleaseStaticSnapshot(release, stagingDir)
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(finalDir, "new.json"))
	require.NoFileExists(t, filepath.Join(finalDir, "old.json"))
	require.FileExists(t, filepath.Join(backupDir, "old.json"))

	require.NoError(t, RollbackActivatedReleaseStaticSnapshot(release, backupDir))
	require.FileExists(t, filepath.Join(finalDir, "old.json"))
	require.NoFileExists(t, filepath.Join(finalDir, "new.json"))

	require.NoError(t, os.MkdirAll(stagingDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(stagingDir, "new.json"), []byte("{}\n"), 0644))
	backupDir, err = ActivateReleaseStaticSnapshot(release, stagingDir)
	require.NoError(t, err)
	require.NoError(t, CommitActivatedReleaseStaticSnapshot(backupDir))
	require.FileExists(t, filepath.Join(finalDir, "new.json"))
	require.NoDirExists(t, backupDir)
	require.NoDirExists(t, filepath.Dir(stagingDir))
}

func TestCleanupReleaseStaticStagingDirPrunesEmptyPluginDir(t *testing.T) {
	oldFactoryRoot := setting.Config.App.FilePath.Factory
	oldRuntimeRoot := setting.Config.App.RuntimeRootPath
	setting.Config.App.FilePath.Factory = t.TempDir()
	setting.Config.App.RuntimeRootPath = ""
	t.Cleanup(func() {
		setting.Config.App.FilePath.Factory = oldFactoryRoot
		setting.Config.App.RuntimeRootPath = oldRuntimeRoot
	})

	release := Release{
		Id:       123,
		PluginId: "TestPlugin",
		Version:  "1.0.0",
	}
	stagingDir := ReleaseStaticStagingDir(release)
	stagingPluginDir := filepath.Dir(stagingDir)
	require.NoError(t, os.MkdirAll(stagingPluginDir, 0755))

	require.NoError(t, CleanupReleaseStaticStagingDir(stagingDir))
	require.NoDirExists(t, stagingPluginDir)
}

func TestEnsureFishTankReleaseStaticSnapshotCopiesRootTemplateFiles(t *testing.T) {
	oldFactoryRoot := setting.Config.App.FilePath.Factory
	oldRuntimeRoot := setting.Config.App.RuntimeRootPath
	oldPluginSourceRoot := setting.Config.App.PluginSourceRoot
	setting.Config.App.FilePath.Factory = t.TempDir()
	setting.Config.App.RuntimeRootPath = ""
	setting.Config.App.PluginSourceRoot = filepath.Join(t.TempDir(), "plugin-source")
	t.Cleanup(func() {
		setting.Config.App.FilePath.Factory = oldFactoryRoot
		setting.Config.App.RuntimeRootPath = oldRuntimeRoot
		setting.Config.App.PluginSourceRoot = oldPluginSourceRoot
	})

	release := Release{
		Id:          910000000000001,
		PluginId:    "FishTank",
		Version:     "1.0.0",
		RuntimeKind: ReleaseRuntimeKindBuiltin,
		TotalSupply: 10000,
		MintPer:     1,
		MintPrice:   "0",
	}
	sourceDir := filepath.Join(setting.Config.App.PluginSourceRoot, release.PluginId, "910000000000001")
	require.NoError(t, os.MkdirAll(filepath.Join(sourceDir, "generated", "fish"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "asset.meta.json"), []byte(`{
  "schema": "senspace.asset-meta.v2",
  "pluginId": "FishTank",
  "collections": [
    {
      "label": "Tank",
      "key": "tank",
      "assetKind": "component",
      "componentRole": "root",
      "metadataRef": "defaultWaterMeta.json#tanks",
      "unitPrice": "5"
    },
    {
      "label": "Fish",
      "key": "fish",
      "assetKind": "component",
      "componentRole": "child",
      "parentKey": "tank",
      "metadataRef": "generated/fish/{tier}.json",
      "traitHashField": "traitHash",
      "tierConfig": {
        "common": { "price": "5", "supply": 1, "mintLimit": 1 },
        "rare": { "price": "50", "supply": 1, "mintLimit": 1 },
        "epic": { "price": "500", "supply": 1, "mintLimit": 1 },
        "legendary": { "price": "-", "supply": 1, "mintLimit": 0 }
      }
    }
  ]
}
`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "defaultWaterMeta.json"), []byte(`{"tanks":[{"id":"tank-1"}]}`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "generated", "fish", "common.json"), []byte(`[{"id":"fish-common-1","tier":"common","traitHash":"hash-common"}]`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "generated", "fish", "rare.json"), []byte(`[{"id":"fish-rare-1","tier":"rare","traitHash":"hash-rare"}]`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "generated", "fish", "epic.json"), []byte(`[{"id":"fish-epic-1","tier":"epic","traitHash":"hash-epic"}]`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "generated", "fish", "legendary.json"), []byte(`[{"id":"fish-legendary-1","tier":"legendary","traitHash":"hash-legendary"}]`), 0644))

	require.NoError(t, EnsureReleaseStaticSnapshot(release))

	finalDir := ReleaseStaticDir(release)
	require.FileExists(t, filepath.Join(finalDir, "asset.meta.json"))
	require.FileExists(t, filepath.Join(finalDir, "defaultWaterMeta.json"))
	require.FileExists(t, filepath.Join(finalDir, "generated", "fish", "common.json"))

	data, err := os.ReadFile(filepath.Join(finalDir, "release.json"))
	require.NoError(t, err)
	var manifest ReleaseStaticManifest
	require.NoError(t, json.Unmarshal(data, &manifest))
	require.Equal(t, "0", manifest.MintPrice)
	require.EqualValues(t, 10000, manifest.TotalSupply)
	require.EqualValues(t, 1, manifest.MintPer)
	require.Contains(t, manifest.TemplateFiles, "defaultWaterMeta.json")
	require.NotEmpty(t, manifest.TemplateFiles["defaultWaterMeta.json"])
}

func TestEnsureReleaseStaticSnapshotSkipsEmptyAssetMetaTemplate(t *testing.T) {
	oldFactoryRoot := setting.Config.App.FilePath.Factory
	oldRuntimeRoot := setting.Config.App.RuntimeRootPath
	oldPluginSourceRoot := setting.Config.App.PluginSourceRoot
	setting.Config.App.FilePath.Factory = t.TempDir()
	setting.Config.App.RuntimeRootPath = ""
	setting.Config.App.PluginSourceRoot = filepath.Join(t.TempDir(), "plugin-source")
	t.Cleanup(func() {
		setting.Config.App.FilePath.Factory = oldFactoryRoot
		setting.Config.App.RuntimeRootPath = oldRuntimeRoot
		setting.Config.App.PluginSourceRoot = oldPluginSourceRoot
	})

	release := Release{
		Id:          62756290892320,
		PluginId:    "62711292247584",
		Version:     "1.0.0",
		RuntimeKind: ReleaseRuntimeKindArtifact,
		TotalSupply: 1000,
		MintPer:     2,
		MintPrice:   "0",
	}
	sourceDir := filepath.Join(setting.Config.App.PluginSourceRoot, release.PluginId, "62756290892320")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "asset.meta.json"), []byte("{\n  \"collections\": [],\n  \"pluginId\": \"62711292247584\",\n  \"schema\": \"senspace.asset-meta.v1\"\n}\n"), 0644))

	require.NoError(t, EnsureReleaseStaticSnapshot(release))

	finalDir := ReleaseStaticDir(release)
	require.NoFileExists(t, filepath.Join(finalDir, "asset.meta.json"))

	data, err := os.ReadFile(filepath.Join(finalDir, "release.json"))
	require.NoError(t, err)
	var manifest ReleaseStaticManifest
	require.NoError(t, json.Unmarshal(data, &manifest))
	require.Empty(t, manifest.AssetMetaUrl)
	require.NotContains(t, manifest.TemplateFiles, "asset.meta.json")
	require.EqualValues(t, 1000, manifest.TotalSupply)
	require.EqualValues(t, 2, manifest.MintPer)
}

func TestStageReleaseStaticSnapshotFallsBackToExistingReleaseFiles(t *testing.T) {
	oldFactoryRoot := setting.Config.App.FilePath.Factory
	oldRuntimeRoot := setting.Config.App.RuntimeRootPath
	oldPluginSourceRoot := setting.Config.App.PluginSourceRoot
	setting.Config.App.FilePath.Factory = t.TempDir()
	setting.Config.App.RuntimeRootPath = ""
	setting.Config.App.PluginSourceRoot = filepath.Join(t.TempDir(), "plugin-source")
	t.Cleanup(func() {
		setting.Config.App.FilePath.Factory = oldFactoryRoot
		setting.Config.App.RuntimeRootPath = oldRuntimeRoot
		setting.Config.App.PluginSourceRoot = oldPluginSourceRoot
	})

	release := Release{
		Id:          910000000000001,
		PluginId:    "FishTank",
		Version:     "1.0.0",
		RuntimeKind: ReleaseRuntimeKindBuiltin,
		TotalSupply: 10000,
		MintPer:     1,
		MintPrice:   "10",
	}

	sourceDir := filepath.Join(setting.Config.App.PluginSourceRoot, release.PluginId, "910000000000001")
	require.NoError(t, os.MkdirAll(sourceDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "asset.meta.json"), []byte(`{
  "schema": "senspace.asset-meta.v2",
  "pluginId": "FishTank",
  "collections": [
    {
      "label": "Tank",
      "key": "tank",
      "assetKind": "component",
      "componentRole": "root",
      "metadataRef": "defaultWaterMeta.json#tanks",
      "unitPrice": "5"
    },
    {
      "label": "Fish",
      "key": "fish",
      "assetKind": "component",
      "componentRole": "child",
      "parentKey": "tank",
      "metadataRef": "generated/fish/{tier}.json",
      "traitHashField": "traitHash",
      "tierConfig": {
        "common": { "price": "5", "supply": 1, "mintLimit": 1 },
        "rare": { "price": "50", "supply": 1, "mintLimit": 1 },
        "epic": { "price": "500", "supply": 1, "mintLimit": 1 },
        "legendary": { "price": "-", "supply": 1, "mintLimit": 0 }
      }
    }
  ]
}
`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "defaultWaterMeta.json"), []byte(`{"tanks":[{"id":"tank-1"}]}`), 0644))

	finalDir := ReleaseStaticDir(release)
	require.NoError(t, os.MkdirAll(filepath.Join(finalDir, "generated", "fish"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(finalDir, "generated", "fish", "common.json"), []byte(`[{"id":"fish-common-1","tier":"common","traitHash":"hash-common"}]`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(finalDir, "generated", "fish", "rare.json"), []byte(`[{"id":"fish-rare-1","tier":"rare","traitHash":"hash-rare"}]`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(finalDir, "generated", "fish", "epic.json"), []byte(`[{"id":"fish-epic-1","tier":"epic","traitHash":"hash-epic"}]`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(finalDir, "generated", "fish", "legendary.json"), []byte(`[{"id":"fish-legendary-1","tier":"legendary","traitHash":"hash-legendary"}]`), 0644))

	stagingDir, err := StageReleaseStaticSnapshot(release)
	require.NoError(t, err)

	require.FileExists(t, filepath.Join(stagingDir, "asset.meta.json"))
	require.FileExists(t, filepath.Join(stagingDir, "defaultWaterMeta.json"))
	require.FileExists(t, filepath.Join(stagingDir, "generated", "fish", "common.json"))
	require.FileExists(t, filepath.Join(stagingDir, "generated", "fish", "rare.json"))
	require.FileExists(t, filepath.Join(stagingDir, "generated", "fish", "epic.json"))
	require.FileExists(t, filepath.Join(stagingDir, "generated", "fish", "legendary.json"))
}

func findRepoRootWithWeb(t *testing.T) string {
	t.Helper()

	wd, err := os.Getwd()
	require.NoError(t, err)
	for {
		webDir := filepath.Join(wd, "..", "senspace-web")
		if info, err := os.Stat(webDir); err == nil && info.IsDir() {
			return wd
		}
		next := filepath.Dir(wd)
		require.NotEqual(t, wd, next, "未找到 senspace-web 仓库目录")
		wd = next
	}
}

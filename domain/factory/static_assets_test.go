package factory

import (
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
}

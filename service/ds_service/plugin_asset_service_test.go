package ds_service

import (
	"os"
	"path/filepath"
	"testing"

	"senspace/domain/ds"
	"senspace/pkg/setting"
)

func TestProcessPluginFileBytesPreservesTextFileMetadata(t *testing.T) {
	data := []byte("[00:01.00]hello world\n")

	processed, err := processPluginFileBytes(data, "lyrics.lrc")
	if err != nil {
		t.Fatalf("process plugin file bytes: %v", err)
	}

	if processed.Kind != defaultFileKind {
		t.Fatalf("expected kind %q, got %q", defaultFileKind, processed.Kind)
	}
	if processed.Mime != "text/plain; charset=utf-8" {
		t.Fatalf("unexpected mime: %q", processed.Mime)
	}
	if processed.Ext != ".lrc" {
		t.Fatalf("expected .lrc ext, got %q", processed.Ext)
	}
	if len(processed.Thumb) != 0 {
		t.Fatalf("expected no thumb for text file")
	}
	if processed.Width != 0 || processed.Height != 0 {
		t.Fatalf("expected zero dimensions, got %dx%d", processed.Width, processed.Height)
	}
	if len(processed.Original) != len(data) {
		t.Fatalf("expected original bytes preserved")
	}
}

func TestBuildDraftArtifactCleanupScopes(t *testing.T) {
	drafts := []ds.PluginInstanceDraft{
		{
			Id:            1,
			DraftId:       "draft-1",
			OwnerKey:      "owner-key",
			OwnerAddress:  "0x123",
			ReleaseId:     99,
			PluginId:      "plugin-1",
			PluginVersion: "1.0.0",
		},
	}

	scopes, draftIDs, err := buildDraftArtifactCleanupScopes(drafts)
	if err != nil {
		t.Fatalf("build cleanup scopes: %v", err)
	}
	if len(scopes) != 1 {
		t.Fatalf("expected 1 scope, got %d", len(scopes))
	}
	if len(draftIDs) != 1 || draftIDs[0] != 1 {
		t.Fatalf("unexpected draft ids: %#v", draftIDs)
	}
	if scopes[0].Kind != ds.PluginAssetScopeDraft {
		t.Fatalf("expected draft scope, got %s", scopes[0].Kind)
	}
	if scopes[0].PluginId != "plugin-1" || scopes[0].DraftId != "draft-1" {
		t.Fatalf("unexpected scope payload: %#v", scopes[0])
	}
}

func TestRemoveDraftArtifactDirsRemovesDraftDirectory(t *testing.T) {
	oldPluginAssets := setting.Config.App.FilePath.PluginAssets
	setting.Config.App.FilePath.PluginAssets = t.TempDir()
	defer func() {
		setting.Config.App.FilePath.PluginAssets = oldPluginAssets
	}()

	scope := ds.PluginAssetScope{
		Kind:          ds.PluginAssetScopeDraft,
		OwnerKey:      "owner-key",
		OwnerAddress:  "0x123",
		ReleaseId:     88,
		DraftId:       "draft-1",
		PluginId:      "plugin-1",
		PluginVersion: "1.0.0",
	}
	targetDir := filepath.Join(ds.PluginAssetInstanceDir(scope), "images", "1")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "original.jpg"), []byte("demo"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	if err := removeDraftArtifactDirs([]ds.PluginAssetScope{scope}); err != nil {
		t.Fatalf("remove draft dirs: %v", err)
	}

	if _, err := os.Stat(ds.PluginAssetInstanceDir(scope)); !os.IsNotExist(err) {
		t.Fatalf("expected draft dir removed, got err=%v", err)
	}
}

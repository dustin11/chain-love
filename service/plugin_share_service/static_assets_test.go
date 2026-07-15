package plugin_share_service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"senspace/domain/ds"
	"senspace/pkg/setting"
)

func TestWriteBackgroundAtomicPublishesReadableFile(t *testing.T) {
	root := useTemporarySharedStaticRoot(t)
	key := "test-background.webp"
	want := []byte("webp-test-content")

	if err := writeBackgroundAtomic(key, want); err != nil {
		t.Fatalf("writeBackgroundAtomic() error = %v", err)
	}

	target := filepath.Join(root, "plugin-shared", key)
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("published content = %q, want %q", got, want)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if gotMode := info.Mode().Perm(); gotMode != 0o644 {
		t.Fatalf("published mode = %04o, want 0644", gotMode)
	}
	assertNoTemporaryBackgrounds(t, filepath.Dir(target))
}

func TestWriteBackgroundAtomicRejectsInvalidKeys(t *testing.T) {
	root := useTemporarySharedStaticRoot(t)
	invalidKeys := []string{"", ".", "..", "../escape.webp", "nested/background.webp", `nested\\background.webp`, "/absolute.webp", " background.webp "}

	for _, key := range invalidKeys {
		t.Run(key, func(t *testing.T) {
			if err := writeBackgroundAtomic(key, []byte("invalid")); err == nil {
				t.Fatal("writeBackgroundAtomic() error = nil, want invalid path error")
			}
		})
	}

	sharedRoot := filepath.Join(root, "plugin-shared")
	entries, err := os.ReadDir(sharedRoot)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("invalid keys published %d entries, want 0", len(entries))
	}
	if _, err := os.Stat(filepath.Join(root, "escape.webp")); !os.IsNotExist(err) {
		t.Fatalf("traversal target Stat() error = %v, want not exist", err)
	}
}

func TestWriteBackgroundAtomicDoesNotPublishOnRenameError(t *testing.T) {
	root := useTemporarySharedStaticRoot(t)
	sharedRoot := filepath.Join(root, "plugin-shared")
	if err := os.MkdirAll(filepath.Join(sharedRoot, "blocked.webp"), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	if err := writeBackgroundAtomic("blocked.webp", []byte("must-not-publish")); err == nil {
		t.Fatal("writeBackgroundAtomic() error = nil, want rename error")
	}
	info, err := os.Stat(filepath.Join(sharedRoot, "blocked.webp"))
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if !info.IsDir() {
		t.Fatal("failed write replaced destination directory")
	}
	assertNoTemporaryBackgrounds(t, sharedRoot)
}

func TestEnsureBackgroundReadableRepairsLegacyMode(t *testing.T) {
	root := useTemporarySharedStaticRoot(t)
	sharedRoot := filepath.Join(root, "plugin-shared")
	if err := os.MkdirAll(sharedRoot, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	target := filepath.Join(sharedRoot, "legacy.webp")
	want := []byte("legacy-background")
	if err := os.WriteFile(target, want, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := os.Chmod(target, 0o600); err != nil {
		t.Fatalf("Chmod(0600) error = %v", err)
	}

	if err := ensureBackgroundReadable("legacy.webp"); err != nil {
		t.Fatalf("ensureBackgroundReadable() error = %v", err)
	}
	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if gotMode := info.Mode().Perm(); gotMode != 0o644 {
		t.Fatalf("repaired mode = %04o, want 0644", gotMode)
	}
	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("repaired content = %q, want %q", got, want)
	}
}

func TestRemoveBackgroundRejectsTraversal(t *testing.T) {
	root := useTemporarySharedStaticRoot(t)
	target := filepath.Join(root, "outside.webp")
	if err := os.WriteFile(target, []byte("keep"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := removeBackground("../outside.webp"); err == nil {
		t.Fatal("removeBackground() error = nil, want invalid path error")
	}
	if _, err := os.Stat(target); err != nil {
		t.Fatalf("outside file was changed: %v", err)
	}
}

func TestRemoveShareResourcesDeletesShareBackgroundOnly(t *testing.T) {
	root := useTemporarySharedStaticRoot(t)
	backgroundKey := "revoke-background.webp"
	if err := writeBackgroundAtomic(backgroundKey, []byte("background")); err != nil {
		t.Fatalf("writeBackgroundAtomic() error = %v", err)
	}

	sourceAsset := filepath.Join(root, "plugin-assets", "source", "asset.png")
	if err := os.MkdirAll(filepath.Dir(sourceAsset), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(sourceAsset, []byte("source-asset"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := removeShareResources(ds.PluginShare{
		BackgroundKey:   backgroundKey,
		ResourceMapJson: `{"asset":{"path":"` + sourceAsset + `"}}`,
	}); err != nil {
		t.Fatalf("removeShareResources() error = %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "plugin-shared", backgroundKey)); !os.IsNotExist(err) {
		t.Fatalf("share background still exists, stat error = %v", err)
	}
	if _, err := os.Stat(sourceAsset); err != nil {
		t.Fatalf("source asset was deleted: %v", err)
	}
}

func useTemporarySharedStaticRoot(t *testing.T) string {
	t.Helper()
	previous := setting.Config.App.RuntimeRootPath
	root := t.TempDir()
	setting.Config.App.RuntimeRootPath = root
	t.Cleanup(func() {
		setting.Config.App.RuntimeRootPath = previous
	})
	return root
}

func assertNoTemporaryBackgrounds(t *testing.T, root string) {
	t.Helper()
	entries, err := os.ReadDir(root)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".plugin-share-") {
			t.Fatalf("temporary file remains: %s", entry.Name())
		}
	}
}

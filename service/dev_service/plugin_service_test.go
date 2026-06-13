package dev_service

import (
	"os"
	"path/filepath"
	"testing"

	"senspace/domain/dev"
	"senspace/domain/ds"
	"senspace/pkg/app/security"
	"senspace/pkg/setting"
)

func TestRemoveDevPluginAssetDirRemovesPluginDirectory(t *testing.T) {
	oldPluginAssets := setting.Config.App.FilePath.PluginAssets
	setting.Config.App.FilePath.PluginAssets = t.TempDir()
	defer func() {
		setting.Config.App.FilePath.PluginAssets = oldPluginAssets
	}()

	ownerKey := "abcdef0123456789"
	pluginID := "62909925865632"
	targetDir := filepath.Join(
		ds.PluginAssetsRoot(),
		string(ds.PluginAssetScopeDev),
		ownerKey,
		pluginID,
		"1.0.0",
		"assets",
		"1001",
	)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		t.Fatalf("mkdir target dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetDir, "original.jpg"), []byte("demo"), 0o644); err != nil {
		t.Fatalf("write test file: %v", err)
	}

	if err := removeDevPluginAssetDir(ownerKey, pluginID); err != nil {
		t.Fatalf("remove dev plugin asset dir: %v", err)
	}

	pluginDir := filepath.Join(
		ds.PluginAssetsRoot(),
		string(ds.PluginAssetScopeDev),
		ownerKey,
		pluginID,
	)
	if _, err := os.Stat(pluginDir); !os.IsNotExist(err) {
		t.Fatalf("expected plugin asset dir to be removed, got err=%v", err)
	}
}

func TestResolveDevPluginOwnerKeyPrefersStoredOwnerKey(t *testing.T) {
	plugin := dev.Plugin{
		OwnerKey: "stored-owner-key",
	}
	user := &security.JwtUser{
		Addr: "0x1234567890abcdef1234567890abcdef12345678",
	}

	ownerKey := resolveDevPluginOwnerKey(plugin, user)

	if ownerKey != "stored-owner-key" {
		t.Fatalf("expected stored owner key, got %q", ownerKey)
	}
}

func TestResolveDevPluginOwnerKeyFallsBackToRequestWallet(t *testing.T) {
	plugin := dev.Plugin{}
	user := &security.JwtUser{
		Addr: "0x1234567890abcdef1234567890abcdef12345678",
	}

	ownerKey := resolveDevPluginOwnerKey(plugin, user)

	if ownerKey == "" {
		t.Fatal("expected fallback owner key to be resolved")
	}
}

package factory_service

import (
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadRuntimeBuildResultValidatesManifest(t *testing.T) {
	outputDir := t.TempDir()
	runtimeDir := filepath.Join(outputDir, "runtime")
	require.NoError(t, os.MkdirAll(runtimeDir, 0o755))

	entryContent := []byte("export default { pluginId: 'demo-plugin' };\n")
	require.NoError(t, os.WriteFile(filepath.Join(runtimeDir, "index.js"), entryContent, 0o644))
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
		PluginId:   req.PluginId,
		Version:    req.Version,
		ReleaseId:  "10001",
		BundleHash: bundleHash,
		Integrity:  integrity,
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
	require.NoError(t, os.WriteFile(filepath.Join(runtimeDir, "index.js"), entryContent, 0o644))
	bundleHash, integrity := runtimeHashes(entryContent)

	req := PluginBuildRequest{
		PluginId:  "demo-plugin",
		Version:   "1.2.3",
		ReleaseId: 10001,
	}

	manifest := runtimeManifestFile{
		PluginId:   req.PluginId,
		Version:    req.Version,
		ReleaseId:  "10001",
		BundleHash: bundleHash,
		Integrity:  integrity,
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

func runtimeHashes(entryContent []byte) (string, string) {
	sha256Sum := sha256.Sum256(entryContent)
	sha384Sum := sha512.Sum384(entryContent)
	return "sha256:" + hex.EncodeToString(sha256Sum[:]),
		"sha384-" + base64.StdEncoding.EncodeToString(sha384Sum[:])
}

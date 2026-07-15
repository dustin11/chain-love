package plugin_share_service

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"senspace/domain/ds"
)

func TestProjectJSONRemovesPrivateScopeAndRewritesRuntimeIDs(t *testing.T) {
	raw := json.RawMessage(`{
		"planetId":10001,
		"ownerKey":"secret",
		"instanceId":"AudioCD-1",
		"attachment":{"surfaceId":"planet-surface"},
		"legacyUrl":"/static/plugin-assets/dev/private/original.png"
	}`)
	projected, err := projectJSON(raw, "AudioCD-1", "planet-surface")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(projected, "10001") || strings.Contains(projected, "secret") {
		t.Fatalf("private identifiers leaked: %s", projected)
	}
	if strings.Contains(projected, "/static/plugin-assets/") {
		t.Fatalf("direct private resource URL leaked: %s", projected)
	}
	if !strings.Contains(projected, "shared-plugin-1") || !strings.Contains(projected, "shared-surface-1") {
		t.Fatalf("share aliases missing: %s", projected)
	}
}

func TestProjectPluginDescriptorPreservesFactoryIdentityForDevShare(t *testing.T) {
	raw := json.RawMessage(`{
		"kind":"local",
		"factoryId":"AudioCD-1",
		"pluginId":"AudioCD-1",
		"version":"dev",
		"options":{"id":"AudioCD-1","surfaceId":"planet-surface","ownerKey":"secret"}
	}`)
	projected, err := projectPluginDescriptor(raw, "AudioCD-1", "planet-surface")
	if err != nil {
		t.Fatal(err)
	}
	var descriptor map[string]any
	if err := json.Unmarshal([]byte(projected), &descriptor); err != nil {
		t.Fatal(err)
	}
	if descriptor["factoryId"] != "AudioCD-1" || descriptor["pluginId"] != "AudioCD-1" {
		t.Fatalf("factory identity was projected as an instance alias: %s", projected)
	}
	options, _ := descriptor["options"].(map[string]any)
	if options["id"] != "shared-plugin-1" || options["surfaceId"] != "shared-surface-1" {
		t.Fatalf("runtime aliases were not projected: %s", projected)
	}
	if _, exists := options["ownerKey"]; exists {
		t.Fatalf("private dev scope leaked: %s", projected)
	}
}

func TestRestoreLegacyPluginDescriptorRepairsSharedFactoryAlias(t *testing.T) {
	restored, err := restoreLegacyPluginDescriptor(
		`{"kind":"local","factoryId":"shared-plugin-1","pluginId":"shared-plugin-1"}`,
		"AudioCD",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(restored, `"factoryId":"AudioCD"`) ||
		!strings.Contains(restored, `"pluginId":"AudioCD"`) {
		t.Fatalf("legacy descriptor was not repaired: %s", restored)
	}
}

func TestOpaqueTokensAreRandomAndOnlyHashesAreStable(t *testing.T) {
	first, err := generateOpaqueToken(32)
	if err != nil {
		t.Fatal(err)
	}
	second, err := generateOpaqueToken(32)
	if err != nil {
		t.Fatal(err)
	}
	if first == second || len(first) < 40 {
		t.Fatalf("tokens are not sufficiently random: %q %q", first, second)
	}
	if hashToken(first) == first || len(hashToken(first)) != 64 {
		t.Fatalf("unexpected token hash: %q", hashToken(first))
	}
}

func TestLegacyShareListItemDoesNotInventShareURL(t *testing.T) {
	item := toShareListItem(ds.PluginShare{
		Id:                   7,
		Status:               ds.PluginShareStatusActive,
		PluginDescriptorJson: `{"pluginId":"AudioCD"}`,
	}, time.Now())

	if item.ShareUrl != "" {
		t.Fatalf("legacy share URL = %q, want empty", item.ShareUrl)
	}
	if item.PluginName != "AudioCD" {
		t.Fatalf("plugin name = %q, want AudioCD", item.PluginName)
	}
}

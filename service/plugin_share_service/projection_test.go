package plugin_share_service

import (
	"encoding/json"
	"strings"
	"testing"
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

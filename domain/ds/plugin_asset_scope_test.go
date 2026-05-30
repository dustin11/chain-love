package ds

import "testing"

func TestValidatePluginAssetPathSegmentRejectsTraversal(t *testing.T) {
	invalidValues := []string{
		"",
		".",
		"..",
		"../demo",
		"demo/1.0.0",
		`demo\1.0.0`,
		"demo plugin",
	}
	for _, value := range invalidValues {
		if err := ValidatePluginAssetPathSegment("segment", value); err == nil {
			t.Fatalf("expected %q to be rejected", value)
		}
	}
}

func TestPluginAssetScopeValidateRejectsUnsafeDevPathParts(t *testing.T) {
	scope := PluginAssetScope{
		Kind:          PluginAssetScopeDev,
		OwnerKey:      "abcdef0123456789",
		PluginId:      "../plugin",
		PluginVersion: "1.0.0",
	}
	if err := scope.Validate(); err == nil {
		t.Fatal("expected unsafe pluginId to be rejected")
	}

	scope.PluginId = "12345"
	scope.PluginVersion = "../1.0.0"
	if err := scope.Validate(); err == nil {
		t.Fatal("expected unsafe pluginVersion to be rejected")
	}
}

func TestPluginAssetScopeStaticPathPartsUseRawSafeSegments(t *testing.T) {
	scope := PluginAssetScope{
		Kind:          PluginAssetScopeDev,
		OwnerKey:      "abcdef0123456789",
		PluginId:      "12345",
		PluginVersion: "1.0.0-beta+1",
	}
	if err := scope.Validate(); err != nil {
		t.Fatalf("scope should be valid: %v", err)
	}
	parts := scope.StaticPathParts()
	if got := parts[len(parts)-1]; got != "1.0.0-beta+1" {
		t.Fatalf("unexpected version path segment: %s", got)
	}
}

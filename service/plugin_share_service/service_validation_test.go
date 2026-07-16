package plugin_share_service

import (
	"encoding/json"
	"testing"
)

// TestValidateCreateJSONAcceptsLocalAudioCDDescriptor 保证本地 AudioCD 分享描述与前端约定一致。
func TestValidateCreateJSONAcceptsLocalAudioCDDescriptor(t *testing.T) {
	input := CreateInput{
		Plugin:        json.RawMessage(`{"kind":"local","factoryId":"AudioCD"}`),
		Carrier:       json.RawMessage(`{"surfaceKind":"desktop"}`),
		Camera:        json.RawMessage(`{"position":{"x":0,"y":0,"z":1}}`),
		State:         json.RawMessage(`{}`),
		ResourceState: json.RawMessage(`{}`),
	}
	if err := validateCreateJSON(input); err != nil {
		t.Fatalf("local AudioCD descriptor must be accepted: %v", err)
	}
}

func TestReplaceResourceAssetIDsRemovesSourceURLsForSharedResources(t *testing.T) {
	state := map[string]any{
		"tracks": []any{map[string]any{
			"audio": map[string]any{
				"assetId":    "42",
				"url":        "blob:http://localhost/source-audio",
				"runtimeUrl": "/static/plugin-assets/source-audio.mp3",
			},
			"title": "保留非资源字段",
		}},
	}

	rewritten := replaceResourceAssetIDs(state, map[uint64]string{42: "shared-asset-1"}).(map[string]any)
	track := rewritten["tracks"].([]any)[0].(map[string]any)
	audio := track["audio"].(map[string]any)
	if got := audio["assetId"]; got != "shared-asset-1" {
		t.Fatalf("assetId = %v, want shared-asset-1", got)
	}
	if _, exists := audio["url"]; exists {
		t.Fatal("shared resource must not retain source url")
	}
	if _, exists := audio["runtimeUrl"]; exists {
		t.Fatal("shared resource must not retain source runtimeUrl")
	}
	if got := track["title"]; got != "保留非资源字段" {
		t.Fatalf("title = %v, want non-resource fields preserved", got)
	}
}

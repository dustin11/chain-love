package terrain

import (
	"encoding/json"
	"strings"
	"testing"

	terrain_domain "senspace/domain/planet/terrain"
	"senspace/pkg/bizerr"
)

// 覆盖 JSON 压缩、非法输入和记录上限。
func TestValidateState(t *testing.T) {
	t.Run("compacts a valid state", func(t *testing.T) {
		state, err := validateState(json.RawMessage(`{
			"platforms": [{
				"id":"p1",
				"kind":"platform",
				"materialId":"grass",
				"transform":{
					"position":[0,6,0],
					"rotation":[0,0,0],
					"scale":[5,1,5]
				},
				"heightField": {
					"version":2,
					"enabled":true,
					"baseHeight":0.006,
					"cellSize":0.5,
					"samplesPerChunk":64,
					"heightUnit":0.005,
					"zeroCode":32768,
					"chunks":[{"x":0,"z":0,"encoding":"constant","code":32868}]
				}
			}],
			"objects": [{
				"id":"o1",
				"kind":"object",
				"platformId":"p1",
				"presetId":"box",
				"materialId":"marble",
				"variantSeed":42,
				"transform":{
					"position":[1,6,1],
					"rotation":[0,0.5,0],
					"scale":[1,1,1]
				}
			}]
		}`))
		if err != nil {
			t.Fatalf("validate state: %v", err)
		}
		if strings.Contains(string(state), "\n") {
			t.Fatalf("expected compact JSON, got %q", string(state))
		}
		if !strings.Contains(string(state), `"materialId":"marble"`) {
			t.Fatalf("expected shape material, got %q", string(state))
		}
		if !strings.Contains(string(state), `"platformId":"p1"`) {
			t.Fatalf("expected platform attachment, got %q", string(state))
		}
	})

	t.Run("rejects an unknown platform attachment", func(t *testing.T) {
		_, err := validateState(json.RawMessage(`{
			"platforms":[],
			"objects":[{
				"id":"o1",
				"kind":"object",
				"platformId":"missing-platform",
				"presetId":"rock",
				"variantSeed":1,
				"transform":{
					"position":[0,0,0],
					"rotation":[0,0,0],
					"scale":[1,1,1]
				}
			}]
		}`))
		if !bizerr.IsKind(err, bizerr.KindParameter) {
			t.Fatalf("expected parameter error, got %v", err)
		}
	})

	t.Run("rejects malformed height chunks", func(t *testing.T) {
		_, err := validateState(json.RawMessage(`{
			"platforms":[],
			"objects":[]
			"heightField":{
				"version":2,
				"enabled":true,
				"baseHeight":4.7,
				"cellSize":0.5,
				"samplesPerChunk":64,
				"heightUnit":0.005,
				"zeroCode":32768,
				"chunks":[{"x":0,"z":0,"encoding":"delta-rle-v1","data":"AQ=="}]
			}
		}`))
		if !bizerr.IsKind(err, bizerr.KindParameter) {
			t.Fatalf("expected parameter error, got %v", err)
		}
	})

	t.Run("accepts a complete compressed height chunk", func(t *testing.T) {
		_, err := validateState(json.RawMessage(`{
			"platforms":[{
				"id":"p1",
				"kind":"platform",
				"materialId":"grass",
				"transform":{"position":[0,6,0],"rotation":[0,0,0],"scale":[5,1,5]},
				"heightField":{
					"version":2,
					"enabled":true,
					"baseHeight":0.006,
					"cellSize":0.5,
					"samplesPerChunk":64,
					"heightUnit":0.005,
					"zeroCode":32768,
					"chunks":[{"x":0,"z":0,"encoding":"delta-rle-v1","data":"AcgB/x8A"}]
				}
				}],
				"objects":[]
			}`))
		if err != nil {
			t.Fatalf("validate compressed chunk: %v", err)
		}
	})

	t.Run("rejects legacy automatic flat height chunks", func(t *testing.T) {
		_, err := validateState(json.RawMessage(`{
			"platforms":[],
			"objects":[],
			"heightField":{
				"version":2,
				"enabled":true,
				"baseHeight":4.7,
				"cellSize":0.5,
				"samplesPerChunk":64,
				"heightUnit":0.005,
				"zeroCode":32768,
				"chunks":[{"x":0,"z":0,"encoding":"constant","code":32768}]
			}
		}`))
		if !bizerr.IsKind(err, bizerr.KindParameter) {
			t.Fatalf("expected parameter error, got %v", err)
		}
	})

	t.Run("rejects textures on non-shape presets", func(t *testing.T) {
		_, err := validateState(json.RawMessage(`{
			"platforms":[],
			"objects":[{
				"id":"o1",
				"kind":"object",
				"presetId":"tulip-patch",
				"materialId":"marble",
				"variantSeed":1,
				"transform":{
					"position":[0,0,0],
					"rotation":[0,0,0],
					"scale":[1,1,1]
				}
			}]
		}`))
		if !bizerr.IsKind(err, bizerr.KindParameter) {
			t.Fatalf("expected parameter error, got %v", err)
		}
	})

	t.Run("rejects unknown presets", func(t *testing.T) {
		_, err := validateState(json.RawMessage(`{
			"platforms":[],
			"objects":[{
				"id":"o1",
				"kind":"object",
				"presetId":"scripted-model",
				"variantSeed":1,
				"transform":{
					"position":[0,0,0],
					"rotation":[0,0,0],
					"scale":[1,1,1]
				}
			}]
		}`))
		if !bizerr.IsKind(err, bizerr.KindParameter) {
			t.Fatalf("expected parameter error, got %v", err)
		}
	})

	t.Run("rejects duplicate ids", func(t *testing.T) {
		_, err := validateState(json.RawMessage(`{
			"platforms":[{
				"id":"same",
				"kind":"platform",
				"materialId":"grass",
				"transform":{
					"position":[0,0,0],
					"rotation":[0,0,0],
					"scale":[1,1,1]
				}
			}],
			"objects":[{
				"id":"same",
				"kind":"object",
				"presetId":"rock",
				"variantSeed":1,
				"transform":{
					"position":[0,0,0],
					"rotation":[0,0,0],
					"scale":[1,1,1]
				}
			}]
		}`))
		if !bizerr.IsKind(err, bizerr.KindParameter) {
			t.Fatalf("expected parameter error, got %v", err)
		}
	})

	t.Run("rejects degenerate transforms", func(t *testing.T) {
		_, err := validateState(json.RawMessage(`{
			"platforms":[],
			"objects":[{
				"id":"o1",
				"kind":"object",
				"presetId":"rock",
				"variantSeed":1,
				"transform":{
					"position":[0,0,0],
					"rotation":[0,0,0],
					"scale":[1,0,1]
				}
			}]
		}`))
		if !bizerr.IsKind(err, bizerr.KindParameter) {
			t.Fatalf("expected parameter error, got %v", err)
		}
	})

	t.Run("rejects non-object state", func(t *testing.T) {
		_, err := validateState(json.RawMessage(`[]`))
		if !bizerr.IsKind(err, bizerr.KindParameter) {
			t.Fatalf("expected parameter error, got %v", err)
		}
	})

	t.Run("rejects too many platforms", func(t *testing.T) {
		platforms := make([]map[string]string, maxPlatforms+1)
		for index := range platforms {
			platforms[index] = map[string]string{"id": "p"}
		}
		raw, err := json.Marshal(map[string]interface{}{
			"platforms": platforms,
			"objects":   []interface{}{},
		})
		if err != nil {
			t.Fatalf("marshal state: %v", err)
		}
		_, err = validateState(raw)
		if !bizerr.IsKind(err, bizerr.KindParameter) {
			t.Fatalf("expected parameter error, got %v", err)
		}
	})
}

// 围栏配置必须同时命中模型和纹理白名单。
func TestValidateFence(t *testing.T) {
	if err := validateFence(&terrainFence{ModelId: "brick-curb", MaterialId: "brick"}); err != nil {
		t.Fatalf("validate fence: %v", err)
	}
	for name, fence := range map[string]*terrainFence{
		"model":    {ModelId: "scripted", MaterialId: "brick"},
		"material": {ModelId: "brick-wall", MaterialId: "plastic"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := validateFence(fence); !bizerr.IsKind(err, bizerr.KindParameter) {
				t.Fatalf("expected parameter error, got %v", err)
			}
		})
	}
}

// 确保历史文档公开读取时升级为平台自带高度场的结构。
func TestMapLegacyDocument(t *testing.T) {
	response := mapDocument(terrain_domain.Document{
		SchemaVersion: 1,
		Revision:      7,
		StateJson:     `{"platforms":[],"objects":[]}`,
		ContentHash:   strings.Repeat("0", 64),
	})
	if response.SchemaVersion != currentSchemaVersion {
		t.Fatalf("expected schema %d, got %d", currentSchemaVersion, response.SchemaVersion)
	}
	if string(response.State) != `{"objects":[],"platforms":[]}` {
		t.Fatalf("expected clean migrated state, got %s", response.State)
	}
	if response.ContentHash == strings.Repeat("0", 64) {
		t.Fatal("expected migrated state hash")
	}
}

// 确保阶段三初版自动高度场读取时被整体清理。
func TestMapAutomaticHeightFieldDocument(t *testing.T) {
	response := mapDocument(terrain_domain.Document{
		SchemaVersion: 2,
		Revision:      8,
		StateJson: `{
			"platforms":[],
			"objects":[],
			"heightField":{
				"version":1,
				"enabled":true,
				"baseHeight":4.7,
				"cellSize":0.5,
				"samplesPerChunk":64,
				"heightUnit":0.005,
				"zeroCode":32768,
				"chunks":[{"x":0,"z":0,"encoding":"constant","code":32868}]
			}
		}`,
		ContentHash: strings.Repeat("0", 64),
	})
	if string(response.State) != `{"objects":[],"platforms":[]}` {
		t.Fatalf("expected removed unowned height field, got %s", response.State)
	}
	if response.ContentHash == strings.Repeat("0", 64) {
		t.Fatal("expected cleaned state hash")
	}
}

// schema 3 已经属于平台的高度场在增加围栏字段时必须完整保留。
func TestMapHeightFieldDocumentToFenceSchema(t *testing.T) {
	response := mapDocument(terrain_domain.Document{
		SchemaVersion: 3,
		Revision:      9,
		StateJson:     `{"platforms":[{"id":"p1","kind":"platform","materialId":"grass","transform":{"position":[0,0,0],"rotation":[0,0,0],"scale":[1,1,1]},"heightField":{"version":2,"enabled":true,"baseHeight":0.006,"cellSize":0.5,"samplesPerChunk":64,"heightUnit":0.005,"zeroCode":32768,"chunks":[{"x":0,"z":0,"encoding":"constant","code":32868}]}}],"objects":[]}`,
		ContentHash:   strings.Repeat("0", 64),
	})
	if response.SchemaVersion != currentSchemaVersion {
		t.Fatalf("expected schema %d, got %d", currentSchemaVersion, response.SchemaVersion)
	}
	if !strings.Contains(string(response.State), `"code":32868`) {
		t.Fatalf("expected preserved height field, got %s", response.State)
	}
	if response.ContentHash == strings.Repeat("0", 64) {
		t.Fatal("expected migrated state hash")
	}
}

// 确保地形发布只级联本次删除的平台，不影响仍存在的平台。
func TestFindRemovedPlatformIds(t *testing.T) {
	removed, err := findRemovedPlatformIds(
		json.RawMessage(`{"platforms":[{"id":"p1"},{"id":"p2"}],"objects":[]}`),
		json.RawMessage(`{"platforms":[{"id":"p2"},{"id":"p3"}],"objects":[]}`),
	)
	if err != nil {
		t.Fatalf("find removed platforms: %v", err)
	}
	if len(removed) != 1 {
		t.Fatalf("expected one removed platform, got %v", removed)
	}
	if _, exists := removed["p1"]; !exists {
		t.Fatalf("expected p1 to be removed, got %v", removed)
	}
}

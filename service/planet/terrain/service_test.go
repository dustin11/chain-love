package terrain

import (
	"encoding/json"
	"strings"
	"testing"

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
				}
			}],
			"objects": [{
				"id":"o1",
				"kind":"object",
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

package road

import (
	"encoding/json"
	"strings"
	"testing"

	"senspace/pkg/bizerr"
)

// 覆盖道路 JSON 压缩、拓扑和严格字段校验。
func TestValidateState(t *testing.T) {
	valid := json.RawMessage(`{
		"nodes":[
			{"id":"n1","anchor":{"kind":"surface","surfaceId":"desktop","surfacePoint":[0,0,0]},"tangentMode":"auto","width":1.2},
			{"id":"n2","anchor":{"kind":"connector","connectorId":"door:outside","side":"outside"},"tangentMode":"manual","tangentIn":[-1,0,0],"width":1.5}
		],
		"edges":[{"id":"e1","fromNodeId":"n1","toNodeId":"n2","styleId":"asphalt","surfaceMode":"bridge","shoulderWidth":0.4,"maxGrade":0.12,"direction":"both","speedLimit":12,"routeModes":["ground","air"]}]
	}`)

	t.Run("compacts a valid graph", func(t *testing.T) {
		state, err := validateState(valid)
		if err != nil {
			t.Fatalf("validate state: %v", err)
		}
		if strings.Contains(string(state), "\n") {
			t.Fatalf("expected compact JSON, got %q", state)
		}
	})

	tests := map[string]string{
		"unknown field":  `{"nodes":[],"edges":[],"mesh":[]}`,
		"removed style":  strings.Replace(string(valid), `"styleId":"asphalt"`, `"styleId":"stone"`, 1),
		"missing node":   `{"nodes":[],"edges":[{"id":"e","fromNodeId":"a","toNodeId":"b","styleId":"asphalt","surfaceMode":"auto","shoulderWidth":0,"maxGrade":0.1,"direction":"both","speedLimit":1,"routeModes":["ground"]}]}`,
		"self loop":      `{"nodes":[{"id":"a","anchor":{"kind":"frame","frameId":"planet","localPoint":[0,0,0]},"tangentMode":"auto","width":1}],"edges":[{"id":"e","fromNodeId":"a","toNodeId":"a","styleId":"asphalt","surfaceMode":"auto","shoulderWidth":0,"maxGrade":0.1,"direction":"both","speedLimit":1,"routeModes":["ground"]}]}`,
		"invalid anchor": `{"nodes":[{"id":"a","anchor":{"kind":"surface","surfaceId":"desk","surfacePoint":[0,0]},"tangentMode":"auto","width":1}],"edges":[]}`,
		"mixed anchor":   `{"nodes":[{"id":"a","anchor":{"kind":"surface","surfaceId":"desk","surfacePoint":[0,0,0],"side":"outside"},"tangentMode":"auto","width":1}],"edges":[]}`,
		"control id":     `{"nodes":[{"id":"a\u0000b","anchor":{"kind":"frame","frameId":"planet","localPoint":[0,0,0]},"tangentMode":"auto","width":1}],"edges":[]}`,
		"duplicate mode": `{"nodes":[{"id":"a","anchor":{"kind":"frame","frameId":"planet","localPoint":[0,0,0]},"tangentMode":"auto","width":1},{"id":"b","anchor":{"kind":"frame","frameId":"planet","localPoint":[1,0,0]},"tangentMode":"auto","width":1}],"edges":[{"id":"e","fromNodeId":"a","toNodeId":"b","styleId":"asphalt","surfaceMode":"auto","shoulderWidth":0,"maxGrade":0.1,"direction":"both","speedLimit":1,"routeModes":["air","air"]}]}`,
	}
	for name, raw := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := validateState(json.RawMessage(raw))
			if !bizerr.IsKind(err, bizerr.KindParameter) {
				t.Fatalf("expected parameter error, got %v", err)
			}
		})
	}
}

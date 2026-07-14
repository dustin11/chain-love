package plugin_share_service

import (
	"encoding/json"
	"strings"
)

var sensitiveProjectionKeys = map[string]struct{}{
	"ownerkey": {}, "owneraddress": {}, "planetid": {}, "spaceid": {},
	"sourceplanetid": {}, "factassetid": {}, "userid": {}, "creatoruserid": {},
	"createdby": {}, "updatedby": {}, "walletaddress": {}, "wallet": {},
	"hostaddr": {}, "addr": {},
}

func projectJSON(raw json.RawMessage, sourceInstanceID string, sourceSurfaceID string) (string, error) {
	if len(raw) == 0 {
		return "{}", nil
	}
	var value any
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", err
	}
	projected := projectValue(value, sourceInstanceID, sourceSurfaceID)
	data, err := json.Marshal(projected)
	return string(data), err
}

func projectValue(value any, sourceInstanceID string, sourceSurfaceID string) any {
	switch typed := value.(type) {
	case []any:
		result := make([]any, 0, len(typed))
		for _, item := range typed {
			result = append(result, projectValue(item, sourceInstanceID, sourceSurfaceID))
		}
		return result
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			normalizedKey := strings.ToLower(strings.NewReplacer("_", "", "-", "", " ", "").Replace(key))
			if _, sensitive := sensitiveProjectionKeys[normalizedKey]; sensitive {
				continue
			}
			if normalizedKey == "instanceid" || normalizedKey == "plugininstanceid" {
				result[key] = "shared-plugin-1"
				continue
			}
			if normalizedKey == "surfaceid" {
				result[key] = "shared-surface-1"
				continue
			}
			result[key] = projectValue(item, sourceInstanceID, sourceSurfaceID)
		}
		return result
	case string:
		if sourceInstanceID != "" && typed == sourceInstanceID {
			return "shared-plugin-1"
		}
		if sourceSurfaceID != "" && typed == sourceSurfaceID {
			return "shared-surface-1"
		}
		if strings.Contains(typed, "/static/plugin-assets/") {
			return ""
		}
		return typed
	default:
		return value
	}
}

func validJSONObject(raw json.RawMessage, label string) error {
	if len(raw) == 0 {
		return nil
	}
	var value map[string]any
	if err := json.Unmarshal(raw, &value); err != nil {
		return &jsonShapeError{label: label}
	}
	return nil
}

type jsonShapeError struct {
	label string
}

func (e *jsonShapeError) Error() string {
	return e.label + "必须是JSON对象"
}

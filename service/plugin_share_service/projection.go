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

// projectPluginDescriptor 投影可变运行参数，但保留加载插件所需的稳定工厂身份。
// 工厂 ID 可能恰好与源实例 ID 相同，不能被普通实例别名规则替换。
func projectPluginDescriptor(raw json.RawMessage, sourceInstanceID string, sourceSurfaceID string) (string, error) {
	projectedJSON, err := projectJSON(raw, sourceInstanceID, sourceSurfaceID)
	if err != nil {
		return "", err
	}
	var source map[string]any
	var projected map[string]any
	if err := json.Unmarshal(raw, &source); err != nil {
		return "", err
	}
	if err := json.Unmarshal([]byte(projectedJSON), &projected); err != nil {
		return "", err
	}
	for _, key := range []string{"kind", "factoryId", "pluginId", "version", "releaseId"} {
		if value, exists := source[key]; exists {
			projected[key] = value
		}
	}
	data, err := json.Marshal(projected)
	return string(data), err
}

// restoreLegacyPluginDescriptor 修复旧分享中被实例别名误替换的工厂身份。
// 只有工厂字段已明确变成公开实例别名时才使用服务端保存的源实例标识。
func restoreLegacyPluginDescriptor(raw string, sourceInstanceID string) (string, error) {
	var descriptor map[string]any
	if err := json.Unmarshal([]byte(raw), &descriptor); err != nil {
		return "", err
	}
	if descriptor["factoryId"] != "shared-plugin-1" {
		return raw, nil
	}
	sourceInstanceID = strings.TrimSpace(sourceInstanceID)
	if sourceInstanceID == "" {
		return raw, nil
	}
	descriptor["factoryId"] = sourceInstanceID
	if descriptor["pluginId"] == "shared-plugin-1" {
		descriptor["pluginId"] = sourceInstanceID
	}
	data, err := json.Marshal(descriptor)
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

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
	preservePluginDescriptorIdentity(source, projected)
	data, err := json.Marshal(projected)
	return string(data), err
}

// preservePluginDescriptorIdentity 保留加载源码所需的稳定身份，同时允许 id、instanceId 等运行别名继续投影。
func preservePluginDescriptorIdentity(source map[string]any, target map[string]any) {
	for _, key := range []string{"kind", "factoryId", "pluginId", "version", "releaseId"} {
		if value, exists := source[key]; exists {
			target[key] = value
		}
	}
	sourceOptions, sourceOK := source["options"].(map[string]any)
	targetOptions, targetOK := target["options"].(map[string]any)
	if !sourceOK || !targetOK {
		return
	}
	for _, key := range []string{"factoryId", "pluginId", "version", "releaseId"} {
		if value, exists := sourceOptions[key]; exists {
			targetOptions[key] = value
		}
	}
}

// remapMomentPluginDescriptor 只替换运行实例/表面引用，不破坏稳定的工厂和源码身份。
func remapMomentPluginDescriptor(raw json.RawMessage, aliases map[string]string) json.RawMessage {
	var source map[string]any
	if json.Unmarshal(raw, &source) != nil {
		return raw
	}
	remapped := remapMomentJSON(raw, aliases)
	var target map[string]any
	if json.Unmarshal(remapped, &target) != nil {
		return raw
	}
	preservePluginDescriptorIdentity(source, target)
	data, err := json.Marshal(target)
	if err != nil {
		return raw
	}
	return data
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
				if text, ok := item.(string); ok && text == sourceInstanceID {
					result[key] = "shared-plugin-1"
				} else {
					result[key] = projectValue(item, sourceInstanceID, sourceSurfaceID)
				}
				continue
			}
			if normalizedKey == "surfaceid" {
				if text, ok := item.(string); ok && text == sourceSurfaceID {
					result[key] = "shared-surface-1"
				} else {
					result[key] = projectValue(item, sourceInstanceID, sourceSurfaceID)
				}
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

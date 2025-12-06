package enum

// AbsEnum 是各具体枚举类型需要实现的接口，保留原有 getDataMap 签名
type AbsEnum interface {
	getDataMap() map[interface{}]string
	// getVal() interface{}
	// getDesc() string
}

// GetList 将任意实现了 AbsEnum 的枚举转换为前端友好的列表结构：
// [{ "val": <value>, "desc": "<description>" }, ...]
func GetList(e AbsEnum) []map[string]interface{} {
	m := e.getDataMap()
	list := make([]map[string]interface{}, 0, len(m))
	for k, v := range m {
		list = append(list, map[string]interface{}{
			"val":  k,
			"desc": v,
		})
	}
	return list
}

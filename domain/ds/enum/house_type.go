package enum

// 房源类型
type HouseType byte

const (
	House_Type_House HouseType = iota + 1
	House_Type_Store
	House_Type_Office
	House_Type_Factory
)

// 独立定义的描述映射，避免每次调用都创建新 map
var houseTypeDesc = map[interface{}]string{
	House_Type_House:   "住宅",
	House_Type_Store:   "商铺",
	House_Type_Office:  "写字楼",
	House_Type_Factory: "厂房",
}

func (v HouseType) getDataMap() map[interface{}]string {
	return houseTypeDesc
}

func (v HouseType) getVal() HouseType {
	return v
}

func (v HouseType) getDesc() string {
	if desc, ok := houseTypeDesc[v]; ok {
		return desc
	}
	return ""
}

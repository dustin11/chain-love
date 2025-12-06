package enum

// 房屋设施
type FurnitureType int16

const (
	Furniture_Refrigerator FurnitureType = 1
	Furniture_TV           FurnitureType = 2
	Furniture_Heater       FurnitureType = 4
)

func (v FurnitureType) getDataMap() map[interface{}]string {
	return map[interface{}]string{
		Furniture_Refrigerator: "冰箱",
		Furniture_TV:           "电视",
		Furniture_Heater:       "热水器",
	}
}

func (v FurnitureType) getVal() FurnitureType {
	return v
}

func (v FurnitureType) getDesc() string {
	return v.getDataMap()[v]
}

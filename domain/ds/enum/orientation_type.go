package enum

// 朝向
type OrientationType int8

const (
	Orientation_East OrientationType = iota + 1
	Orientation_South
	Orientation_West
	Orientation_North
	Orientation_Southeast
	Orientation_Northeast
	Orientation_Southwest
	Orientation_Northwest
	Orientation_North_South
	Orientation_West_East
)

func (v OrientationType) getDataMap() map[interface{}]string {
	return map[interface{}]string{
		Orientation_East:        "东",
		Orientation_South:       "南",
		Orientation_West:        "西",
		Orientation_North:       "北",
		Orientation_Southeast:   "东南",
		Orientation_Northeast:   "东北",
		Orientation_Southwest:   "西南",
		Orientation_Northwest:   "西北",
		Orientation_North_South: "南北",
		Orientation_West_East:   "东西",
	}
}

func (v OrientationType) getVal() OrientationType {
	return v
}

func (v OrientationType) getDesc() string {
	return v.getDataMap()[v]
}

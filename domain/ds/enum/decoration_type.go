package enum

// var initOnce sync.Once
// 装修类型
type DecorationType int8

const (
	Decoration_Rough DecorationType = iota + 1
	Decoration_Simple
	Decoration_Refined
	Decoration_Luxury
)

func (v DecorationType) getDataMap() map[interface{}]string {
	return map[interface{}]string{
		Decoration_Rough:   "毛坯",
		Decoration_Simple:  "简单装修",
		Decoration_Refined: "精装修",
		Decoration_Luxury:  "豪华装修",
	}
}

func (v DecorationType) getVal() DecorationType {
	return v
}

func (v DecorationType) getDesc() string {
	return v.getDataMap()[v]
}

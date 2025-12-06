package enum

// 房源状态
type RentState byte

const (
	Rent_State_NoPublish RentState = iota + 1
	Rent_State_Publish
	Rent_State_Rent
	Rent_State_RentBack
)

func (v RentState) getDataMap() map[interface{}]string {
	return map[interface{}]string{
		Rent_State_NoPublish: "未发布",
		Rent_State_Publish:   "已发布",
		Rent_State_Rent:      "出租中",
		Rent_State_RentBack:  "已退房",
	}
}

func (v RentState) getVal() RentState {
	return v
}

func (v RentState) getDesc() string {
	return v.getDataMap()[v]
}

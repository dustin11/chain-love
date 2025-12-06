package enum

// 出租类型
type RentType byte

const (
	Rent_Type_House RentType = iota + 1
	Rent_Type_Room
	Rent_Type_Store
)

func (v RentType) getDataMap() map[interface{}]string {
	return map[interface{}]string{
		Rent_Type_House: "整租",
		Rent_Type_Room:  "单间",
	}
}

func (v RentType) getVal() RentType {
	return v
}

func (v RentType) getDesc() string {
	return v.getDataMap()[v]
}

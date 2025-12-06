package enum

// 付款方式
type RentPayType byte

const (
	Rent_Pay_1To1 RentPayType = iota + 1
	Rent_Pay_2To1
	Rent_Pay_3To1
	Rent_Pay_HalfYear
	Rent_Pay_Year
)

func (v RentPayType) getDataMap() map[interface{}]string {
	return map[interface{}]string{
		Rent_Pay_1To1:     "押一付一",
		Rent_Pay_2To1:     "押一付二",
		Rent_Pay_3To1:     "押一付二",
		Rent_Pay_HalfYear: "半年付",
		Rent_Pay_Year:     "年付",
	}
}

func (v RentPayType) getVal() RentPayType {
	return v
}

func (v RentPayType) getDesc() string {
	return v.getDataMap()[v]
}

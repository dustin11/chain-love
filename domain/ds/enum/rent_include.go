package enum

// 租金包含
type RentInclude int16

const (
	Rent_Include_Water    RentInclude = 1
	Rent_Include_Electric RentInclude = 2
	Rent_Include_gas      RentInclude = 4
	Rent_Include_Internet RentInclude = 8
	Rent_Include_Property RentInclude = 16
	Rent_Include_Heating  RentInclude = 32
	Rent_Include_CableTV  RentInclude = 64
	Rent_Include_Parking  RentInclude = 128
)

func (v RentInclude) getDataMap() map[interface{}]string {
	return map[interface{}]string{
		Rent_Include_Water:    "水费",
		Rent_Include_Electric: "电费",
		Rent_Include_gas:      "燃气费",
		Rent_Include_Internet: "宽带费",
		Rent_Include_Property: "物业费",
		Rent_Include_Heating:  "取暖费",
		Rent_Include_CableTV:  "有线电视费",
		Rent_Include_Parking:  "停车费",
	}
}

func (v RentInclude) getVal() RentInclude {
	return v
}

func (v RentInclude) getDesc() string {
	return v.getDataMap()[v]
}

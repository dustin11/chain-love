package enum

type AccountType byte

const (
	Account_Type_Email AccountType = iota + 1
	Account_Type_Mobile
)

func (v AccountType) getDataMap() map[interface{}]string {
	return map[interface{}]string{
		Account_Type_Email:  "邮箱",
		Account_Type_Mobile: "手机",
	}
}

func (v AccountType) getVal() AccountType {
	return v
}

func (v AccountType) getDesc() string {
	return v.getDataMap()[v]
}

package enum

// 出租类型
type RowState byte

const (
	Row_State_Draft RowState = iota + 1
	Row_State_Effective
	Row_State_Invalid
)

func (v RowState) getDataMap() map[interface{}]string {
	return map[interface{}]string{
		Row_State_Draft:     "草稿",
		Row_State_Effective: "有效",
		Row_State_Invalid:   "无效",
	}
}

func (v RowState) getVal() RowState {
	return v
}

func (v RowState) getDesc() string {
	return v.getDataMap()[v]
}

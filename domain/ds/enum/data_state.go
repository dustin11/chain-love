package enum

import "fmt"

type DataState byte

const (
	Data_Draft     DataState = 1
	Data_Available DataState = 3
	Data_Delete    DataState = 5
)

var dataStateDesc = map[DataState]string{
	Data_Draft:     "草稿",
	Data_Available: "可用",
	Data_Delete:    "删除或关闭",
}

func (v DataState) String() string {
	if s, ok := dataStateDesc[v]; ok {
		return s
	}
	return fmt.Sprintf("DataState(%d)", v)
}

// 如果需要保留原来方法名
func (v DataState) getDataMap() map[DataState]string {
	return dataStateDesc
}

func (v DataState) getVal() DataState {
	return v
}

func (v DataState) getDesc() string {
	return v.String()
}

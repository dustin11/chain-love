package enum

import "fmt"

type ImgBizType uint8

const (
	Avatar     ImgBizType = 1
	Album      ImgBizType = 5
	DeskSiding ImgBizType = 10
)

var imgBizTypeDesc = map[ImgBizType]string{
	Avatar:     "头像",
	Album:      "相册",
	DeskSiding: "桌靠",
}

func (v ImgBizType) String() string {
	if s, ok := imgBizTypeDesc[v]; ok {
		return s
	}
	return fmt.Sprintf("ImgBizType(%d)", v)
}

// 如果需要保留原来方法名
func (v ImgBizType) getDataMap() map[ImgBizType]string {
	return imgBizTypeDesc
}

func (v ImgBizType) getVal() ImgBizType {
	return v
}

func (v ImgBizType) getDesc() string {
	return v.String()
}

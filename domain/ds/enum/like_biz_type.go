package enum

import "fmt"

type LikeBizType uint8

const (
	DeskSidingAlbum LikeBizType = 1
	DeskSidingNote  LikeBizType = 3
)

var likeBizTypeDesc = map[LikeBizType]string{
	DeskSidingAlbum: "桌靠相册",
	DeskSidingNote:  "桌靠笔记",
}

func (v LikeBizType) String() string {
	if s, ok := likeBizTypeDesc[v]; ok {
		return s
	}
	return fmt.Sprintf("LikeBizType(%d)", v)
}

// 如果需要保留原来方法名
func (v LikeBizType) getDataMap() map[LikeBizType]string {
	return likeBizTypeDesc
}

func (v LikeBizType) getVal() LikeBizType {
	return v
}

func (v LikeBizType) getDesc() string {
	return v.String()
}

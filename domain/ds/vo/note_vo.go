package vo

type NoteStyleVO struct {
	FontSize   string     `json:"fontSize"`
	FontFamily string     `json:"fontFamily"`
	BgColor    string     `json:"bgColor"`
	Pos        [4]float64 `json:"pos"`
}

type NoteVO struct {
	Id        uint64       `json:"id"`
	Text      string       `json:"text"`
	Style     *NoteStyleVO `json:"style,omitempty"`
	LikeCount *int         `json:"likeCount,omitempty"`
}

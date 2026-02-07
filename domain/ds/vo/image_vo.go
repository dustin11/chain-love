package vo

type ImageVO struct {
	Id    uint64 `json:"id"`
	Url   string `json:"url"`
	Style string `json:"style,omitempty"`
}

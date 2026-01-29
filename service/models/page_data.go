package models

// PageData maps to frontend PageData
type PageData struct {
	Idx              int     `json:"idx"`
	Content          string  `json:"content"`
	Title            *string `json:"title,omitempty"`
	ChapterTitle     *string `json:"chapterTitle,omitempty"`
	ChapterType      string  `json:"chapterType,omitempty"`      // "main" | "other"
	ChapterPageIndex *int    `json:"chapterPageIndex,omitempty"` // optional
}

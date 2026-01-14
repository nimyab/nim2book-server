package models

type ChapterAlignNode struct {
	Id              string               `json:"id"`
	Title           string               `json:"title"`
	TranslatedTitle string               `json:"translatedTitle"`
	Content         []ParagraphAlignNode `json:"content"`
	Order           int                  `json:"order"`
}

type ParagraphAlignNode struct {
	OriginalParagraph   string          `json:"op"`
	TranslatedParagraph string          `json:"tp"`
	AlignmentWords      []WordAlignNode `json:"aw"`
}

type WordAlignNode struct {
	IndexesOriginalWord   [2]int `json:"iow"`
	IndexesTranslatedWord [2]int `json:"itw"`
}

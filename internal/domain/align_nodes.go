package domain

type ChapterAlignNode struct {
	Id              string        `json:"id"`
	Title           string        `json:"title"`
	TranslatedTitle string        `json:"translatedTitle"`
	Content         []ContentNode `json:"content"`
	Order           int           `json:"order"`
}

type ContentNode struct {
	Paragraph *ParagraphAlignNode `json:"paragraph,omitempty"`
	Image     *ImageNode          `json:"image,omitempty"`
}

type ImageNode struct {
	Path string `json:"path"`
	Alt  string `json:"alt"`
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

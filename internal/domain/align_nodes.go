package domain

type ChapterAlignNode struct {
	Id              string        `json:"id"`
	Title           string        `json:"title"`
	TranslatedTitle string        `json:"translatedTitle"`
	Content         []ContentNode `json:"content"`
	Order           int           `json:"order"`
}

type ParagraphAlignNodeType string

const (
	ParagraphAlignNodeTypeParagraph ParagraphAlignNodeType = "paragraph"
	ParagraphAlignNodeTypeImage     ParagraphAlignNodeType = "image"
)

type ContentNode struct {
	Type               ParagraphAlignNodeType `json:"type"`
	ImageNode          *ImageNode             `json:"image,omitempty"`
	ParagraphAlignNode *ParagraphAlignNode    `json:"paragraph,omitempty"`
}

type ImageNode struct {
	ImageURL string `json:"img"`
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

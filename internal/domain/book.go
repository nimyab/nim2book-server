package domain

import "time"

type Book struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Title          string        `json:"title"`
	CoverURL       *string       `json:"coverUrl"`
	OriginalLang   string        `json:"originalLang"`
	TranslatedLang string        `json:"translatedLang"`
	ProcessStatus  ProcessStatus `json:"processStatus"`

	Author       *Author        `json:"author"`
	Genres       []*Genre       `json:"genres"`
	BookChapters []*BookChapter `json:"bookChapters"`
}

type BookChapter struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Order           int    `json:"order"`
	Title           string `json:"title"`
	TranslatedTitle string `json:"translatedTitle"`
	ContentURL      string `json:"contentURL"`

	Book *Book `json:"book,omitempty"`
}

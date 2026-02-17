package domain

import "time"

type Book struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Title          string  `json:"title"`
	CoverURL       *string `json:"coverUrl"`
	OriginalLang   string  `json:"originalLang"`
	TranslatedLang string  `json:"translatedLang"`

	Author       *Author        `json:"author"`
	Genres       []*Genre       `json:"genres"`
	BookChapters []*BookChapter `json:"bookChapters"`
}

package domain

import "time"

type BookChapter struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Order           int    `json:"order"`
	Title           string `json:"title"`
	TranslatedTitle string `json:"translatedTitle"`
	ContentURL      string `json:"contentURL"`

	Book *Book `json:"book,omitempty"`
}

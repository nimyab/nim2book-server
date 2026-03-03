package domain

import (
	"time"
)

type PersonalBook struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Title          string        `json:"title"`
	CoverURL       *string       `json:"coverUrl"`
	OriginalLang   string        `json:"originalLang"`
	TranslatedLang string        `json:"translatedLang"`
	ProcessStatus  ProcessStatus `json:"processStatus"`

	User                 *User                  `json:"user,omitempty"`
	Author               *Author                `json:"author"`
	Genres               []*Genre               `json:"genres"`
	PersonalBookChapters []*PersonalBookChapter `json:"personalBookChapters"`
}

type PersonalBookChapter struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Order           int    `json:"order"`
	Title           string `json:"title"`
	TranslatedTitle string `json:"translatedTitle"`
	ContentURL      string `json:"contentUrl"`

	PersonalBook *PersonalBook `json:"personalBook,omitempty"`
}

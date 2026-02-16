package domain

import (
	"time"
)

type PersonalBook struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Title          string `json:"title"`
	CoverURL       string `json:"coverUrl"`
	OriginalLang   string `json:"originalLang"`
	TranslatedLang string `json:"translatedLang"`

	User                 *User                  `json:"user,omitempty"`
	Author               *Author                `json:"author"`
	Genres               []*Genre               `json:"genres"`
	PersonalBookChapters []*PersonalBookChapter `json:"personalBookChapters"`
}

package domain

import (
	"time"
)

type ProcessStatus string

const (
	ProcessStatusNotStarted ProcessStatus = "not_started"
	ProcessStatusInProgress ProcessStatus = "in_progress"
	ProcessStatusCompleted  ProcessStatus = "completed"
	ProcessStatusFailed     ProcessStatus = "failed"
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

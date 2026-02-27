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

func (b *Book) GetID() ID {
	return b.ID
}

func (b *Book) GetProcessStatus() ProcessStatus {
	return b.ProcessStatus
}

func (b *Book) SetProcessStatus(status ProcessStatus) {
	b.ProcessStatus = status
}

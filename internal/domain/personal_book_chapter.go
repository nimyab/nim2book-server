package domain

import (
	"time"
)

type PersonalBookChapter struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Order           int    `json:"order"`
	Title           string `json:"title"`
	TranslatedTitle string `json:"translatedTitle"`
	ContentURL      string `json:"contentUrl"`

	PersonalBook *PersonalBook `json:"personalBook,omitempty"`
}

func (c *PersonalBookChapter) GetOrder() int {
	return c.Order
}

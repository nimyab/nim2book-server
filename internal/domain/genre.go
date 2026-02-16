package domain

import "time"

type Genre struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Name string `json:"name"`

	Books         []*Book         `json:"books,omitempty"`
	PersonalBooks []*PersonalBook `json:"personalBooks,omitempty"`
}

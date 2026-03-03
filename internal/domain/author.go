package domain

import "time"

type Author struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Name string `json:"name"`

	PersonalBooks []*PersonalBook `json:"personalBooks,omitempty"`
	Books         []*Book         `json:"books,omitempty"`
}

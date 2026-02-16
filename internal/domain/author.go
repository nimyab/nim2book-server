package domain

import "time"

type Author struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"created_at"`

	Name string `json:"name"`

	PersonalBooks []*PersonalBook `json:"personal_books,omitempty"`
	Books         []*Book         `json:"books,omitempty"`
}

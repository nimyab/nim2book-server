package domain

import "github.com/google/uuid"

type Book struct {
	Id           uuid.UUID `json:"id" db:"id"`
	Title        string    `json:"title" db:"title"`
	Author       string    `json:"author" db:"author"`
	ChapterPaths []string  `json:"chapter_paths" db:"chapter_paths"`
}

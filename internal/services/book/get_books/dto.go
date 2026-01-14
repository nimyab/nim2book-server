package get_books

import "github.com/nimyab/nim2book-back/internal/models"

type Input struct {
	Author string `query:"author" validate:"omitempty"`
	Title  string `query:"title" validate:"omitempty"`
	Page   int    `query:"page" validate:"required,min=1"`
}

type Output struct {
	Books []models.Book `json:"books"`
}

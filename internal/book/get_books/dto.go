package get_books

import "github.com/nimyab/nim2book-back/internal/domain"

type Input struct {
	Author string `query:"author" validate:"omitempty"`
	Title  string `query:"title" validate:"omitempty"`
	Page   int    `query:"page" validate:"required,min=1"`
}

type Output struct {
	Books []domain.Book `json:"books"`
}

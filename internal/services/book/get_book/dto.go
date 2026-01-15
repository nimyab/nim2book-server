package get_book

import (
	"github.com/nimyab/nim2book-back/internal/domain"
)

type Input struct {
	Id domain.Id `param:"id" validate:"required,uuid"`
}

type Output struct {
	Book *domain.Book `json:"book"`
}

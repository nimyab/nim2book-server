package get_book

import (
	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
)

type Input struct {
	Id uuid.UUID `param:"id" validate:"required,uuid"`
}

type Output struct {
	Book *domain.Book `json:"book"`
}

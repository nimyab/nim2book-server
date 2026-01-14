package get_book

import (
	"github.com/nimyab/nim2book-back/internal/models"
)

type Input struct {
	Id models.ID `param:"id" validate:"required,uuid"`
}

type Output struct {
	Book *models.Book `json:"book"`
}

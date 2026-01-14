package update_book

import "github.com/nimyab/nim2book-back/internal/models"

type Input struct {
	Id     models.ID `param:"id" validate:"required,uuid"`
	Author *string   `form:"author"`
	Title  *string   `form:"title"`
}

type Output struct {
	Book *models.Book `json:"book"`
}

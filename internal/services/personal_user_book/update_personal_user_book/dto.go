package update_personal_user_book

import "github.com/nimyab/nim2book-back/internal/domain"

type Input struct {
	UserId domain.ID `validate:"required,uuid"`
	Id     domain.ID `param:"id" validate:"required,uuid"`
	Author *string   `form:"author"`
	Title  *string   `form:"title"`
}

type Output struct {
	Book *domain.PersonalBook `json:"book"`
}

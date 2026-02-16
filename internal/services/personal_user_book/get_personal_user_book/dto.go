package get_personal_user_book

import "github.com/nimyab/nim2book-back/internal/domain"

type Input struct {
	UserId domain.ID `validate:"required,uuid"`
	BookId domain.ID `param:"id" validate:"required,uuid"`
}

type Output struct {
	Book *domain.PersonalBook `json:"book"`
}

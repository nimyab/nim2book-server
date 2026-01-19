package get_personal_user_book

import "github.com/nimyab/nim2book-back/internal/domain"

type Input struct {
	UserId domain.Id `validate:"required,uuid"`
	BookId domain.Id `param:"id" validate:"required,uuid"`
}

type Output struct {
	Book *domain.PersonalUserBook `json:"book"`
}

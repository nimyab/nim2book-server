package translate_personal_user_book

import "github.com/nimyab/nim2book-back/internal/domain"

type Input struct {
	From domain.SupportedLang `json:"from" form:"from" validate:"required"`
	To   domain.SupportedLang `json:"to" form:"to" validate:"required"`
}

type Output struct {
	Book    *domain.PersonalUserBook `json:"book,omitempty"`
	Message string                   `json:"messageAboutTranslate,omitempty"`
}

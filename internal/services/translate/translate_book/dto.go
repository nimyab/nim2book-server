package translate_book

import "github.com/nimyab/nim2book-back/internal/domain"

type Input struct {
	From domain.SupportedLang `json:"from" form:"from" validate:"required"`
	To   domain.SupportedLang `json:"to" form:"to" validate:"required"`
}

type Output struct {
	Book    *domain.Book `json:"book,omitempty"`
	Message string       `json:"messageAboutTranslate,omitempty"`
}

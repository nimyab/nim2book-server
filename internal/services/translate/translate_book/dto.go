package translate_book

import "github.com/nimyab/nim2book-back/internal/models"

type Input struct {
	From models.SupportedLang `json:"from" form:"from" validate:"required"`
	To   models.SupportedLang `json:"to" form:"to" validate:"required"`
}

type Output struct {
	Book    *models.Book `json:"book,omitempty"`
	Message string       `json:"messageAboutTranslate,omitempty"`
}

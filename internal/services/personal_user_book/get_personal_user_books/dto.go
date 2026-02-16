package get_personal_user_books

import "github.com/nimyab/nim2book-back/internal/domain"

type Input struct {
	UserId  domain.ID  `validate:"required,uuid"`
	Author  string     `query:"author" validate:"omitempty"`
	Title   string     `query:"title" validate:"omitempty"`
	GenreId *domain.ID `query:"genreId" validate:"omitempty,uuid"`
	Page    int        `query:"page" validate:"required,min=1"`
}

type Output struct {
	Books []domain.PersonalBook `json:"books"`
}

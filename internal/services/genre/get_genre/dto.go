package get_genre

import "github.com/nimyab/nim2book-back/internal/domain"

type Input struct {
	Id domain.ID `param:"id" validate:"required,uuid"`
}

type Output struct {
	Genre *domain.Genre `json:"genre"`
}

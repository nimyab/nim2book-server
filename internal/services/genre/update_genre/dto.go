package update_genre

import "github.com/nimyab/nim2book-back/internal/domain"

type Input struct {
	Id   domain.ID `param:"id" validate:"required,uuid"`
	Name string    `json:"name" validate:"required,min=1,max=100"`
}

type Output struct {
	Genre *domain.Genre `json:"genre"`
}

package create_genre

import "github.com/nimyab/nim2book-back/internal/domain"

type Input struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type Output struct {
	Genre *domain.Genre `json:"genre"`
}

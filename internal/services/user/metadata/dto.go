package metadata

import "github.com/nimyab/nim2book-back/internal/domain"

type Input struct {
	UserId   domain.ID    `json:"-" validate:"required"`
	Metadata domain.JsonB `json:"metadata" validate:"required"`
}

type Output struct {
	User *domain.User `json:"user"`
}

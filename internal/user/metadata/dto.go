package metadata

import "github.com/nimyab/nim2book-back/internal/domain"

type Input struct {
	UserId   domain.Id      `json:"-" validate:"required"`
	Metadata map[string]any `json:"metadata" validate:"required"`
}

type Output struct {
	User *domain.User `json:"user"`
}

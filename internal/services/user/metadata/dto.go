package metadata

import "github.com/nimyab/nim2book-back/internal/models"

type Input struct {
	UserId   models.ID    `json:"-" validate:"required"`
	Metadata models.JSONB `json:"metadata" validate:"required"`
}

type Output struct {
	User *models.User `json:"user"`
}

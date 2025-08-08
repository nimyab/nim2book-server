package me

import (
	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
)

type Input struct {
	UserId uuid.UUID `validate:"required,uuid"`
}

type Output struct {
	User *domain.User `json:"user"`
}

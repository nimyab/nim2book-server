package me

import (
	"github.com/nimyab/nim2book-back/internal/domain"
)

type Input struct {
	UserId domain.ID `validate:"required,uuid"`
}

type Output struct {
	User *domain.User `json:"user"`
}

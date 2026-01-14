package me

import (
	"github.com/nimyab/nim2book-back/internal/models"
)

type Input struct {
	UserId models.ID `validate:"required,uuid"`
}

type Output struct {
	User *models.User `json:"user"`
}

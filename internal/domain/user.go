package domain

import (
	"github.com/google/uuid"
)

type User struct {
	Id           uuid.UUID `json:"id" db:"id"`
	Login        string    `json:"login" db:"login"`
	Email        string    `json:"email" db:"email"`
	PasswordHash string    `json:"-" db:"password_hash"`
	Books        []Book    `json:"books"`
}

package domain

import "github.com/google/uuid"

type JwtPayload struct {
	Id      uuid.UUID `json:"id"`
	IsAdmin bool      `json:"isAdmin"`
}

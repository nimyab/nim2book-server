package domain

import (
	"time"
)

type FcmToken struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Token string `json:"token"`

	User *User `json:"user,omitempty"`
}

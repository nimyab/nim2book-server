package domain

import "time"

type BasicAccount struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"created_at"`

	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	IsVerified   bool   `json:"isVerified"`
	VerifyLink   string `json:"verifyLink"`

	User *User `json:"user,omitempty"`
}

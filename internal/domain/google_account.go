package domain

import "time"

type GoogleAccount struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`

	User *User `json:"user,omitempty"`
}

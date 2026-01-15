package domain

import "time"

type FcmToken struct {
	Token    string    `json:"token" db:"token"`
	UserId   Id        `json:"userId" db:"user_id"`
	CreateAt time.Time `json:"createAt" db:"create_at"`
}

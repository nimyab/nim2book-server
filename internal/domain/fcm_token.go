package domain

import "time"

type FcmToken struct {
	Token    string    `json:"token"`
	UserId   Id        `json:"userId"`
	CreateAt time.Time `json:"createAt"`
}

package models

import (
	"time"
)

// FcmToken represents Firebase Cloud Messaging tokens for push notifications
type FcmToken struct {
	Token    string    `gorm:"column:token;type:varchar(255);primaryKey" json:"token"`
	UserID   ID        `gorm:"column:user_id;type:uuid;not null;index" json:"userId"`
	CreateAt time.Time `gorm:"column:create_at;type:timestamp with time zone;default:CURRENT_TIMESTAMP" json:"createAt"`

	// Relations
	User *User `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

func (FcmToken) TableName() string {
	return "fcm_tokens"
}

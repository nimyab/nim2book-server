package models

import (
	"github.com/google/uuid"
)

// User represents application user with multiple auth methods
type User struct {
	BaseModel

	IsAdmin                bool    `gorm:"column:is_admin;default:false;not null" json:"isAdmin"`
	IsVip                  bool    `gorm:"column:is_vip;default:false;not null" json:"isVip"`
	Metadata               JSONB   `gorm:"column:metadata;type:jsonb;default:'{}';not null" json:"metadata"`
	GoogleAccountSub       *string `gorm:"column:google_account_sub;type:varchar(40);uniqueIndex" json:"googleAccountSub,omitempty"`
	EmailPasswordAccountID *ID     `gorm:"column:email_password_account_id;type:uuid;uniqueIndex" json:"emailPasswordAccountId,omitempty"`

	// Relations
	GoogleAccount        *GoogleAccount        `gorm:"foreignKey:GoogleAccountSub;references:Sub" json:"googleAccount,omitempty"`
	EmailPasswordAccount *EmailPasswordAccount `gorm:"foreignKey:EmailPasswordAccountID;references:ID" json:"emailPasswordAccount,omitempty"`
	FcmTokens            []FcmToken            `gorm:"foreignKey:UserID;references:ID;constraint:OnDelete:CASCADE" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// IsGoogleUser checks if user uses Google authentication
func (u User) IsGoogleUser() bool {
	return u.GoogleAccountSub != nil && *u.GoogleAccountSub != ""
}

// IsEmailPasswordUser checks if user uses email/password authentication
func (u User) IsEmailPasswordUser() bool {
	return u.EmailPasswordAccountID != nil && *u.EmailPasswordAccountID != uuid.Nil
}

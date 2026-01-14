package models

// GoogleAccount represents Google OAuth account
type GoogleAccount struct {
	Sub           string  `gorm:"column:sub;type:varchar(40);primaryKey" json:"sub"`
	Email         string  `gorm:"column:email;type:varchar(255);not null" json:"email"`
	EmailVerified bool    `gorm:"column:email_verified;not null" json:"emailVerified"`
	Name          string  `gorm:"column:name;type:varchar(255);not null" json:"name"`
	Picture       *string `gorm:"column:picture;type:varchar(255)" json:"picture,omitempty"`

	// Relations
	User *User `gorm:"foreignKey:GoogleAccountSub;references:Sub;constraint:OnDelete:CASCADE" json:"user,omitempty"`
}

func (GoogleAccount) TableName() string {
	return "google_accounts"
}

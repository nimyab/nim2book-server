package models

// EmailPasswordAccount represents email/password authentication
type EmailPasswordAccount struct {
	BaseModel

	Email        string `gorm:"column:email;type:varchar(255);not null;uniqueIndex" json:"email"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(255);not null" json:"-"`

	// Relations
	User *User `gorm:"foreignKey:EmailPasswordAccountID;references:ID" json:"user,omitempty"`
}

func (EmailPasswordAccount) TableName() string {
	return "email_password_accounts"
}

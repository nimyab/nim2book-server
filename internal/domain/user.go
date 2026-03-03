package domain

import "time"

type User struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	IsAdmin  bool  `json:"isAdmin"`
	IsVIP    bool  `json:"isVIP"`
	Metadata JsonB `json:"metadata"`

	GoogleAccount *GoogleAccount  `json:"googleAccount,omitempty"`
	BasicAccount  *BasicAccount   `json:"basicAccount,omitempty"`
	PersonalBooks []*PersonalBook `json:"personalBooks"`
	FcmTokens     []*FcmToken     `json:"-"`
}

type GoogleAccount struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"emailVerified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`

	User *User `json:"-"`
}

type BasicAccount struct {
	ID        ID        `json:"id"`
	CreatedAt time.Time `json:"createdAt"`

	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	IsVerified   bool   `json:"isVerified"`
	VerifyLink   string `json:"verifyLink"`

	User *User `json:"-"`
}

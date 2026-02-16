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

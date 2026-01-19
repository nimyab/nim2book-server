package domain

type User struct {
	Id       Id    `json:"id"`
	IsAdmin  bool  `json:"isAdmin"`
	IsVIP    bool  `json:"isVIP"`
	Metadata JsonB `json:"metadata"`

	GoogleAccount        *GoogleAccount        `json:"googleAccount,omitempty"`
	EmailPasswordAccount *EmailPasswordAccount `json:"emailPasswordAccount,omitempty"`
	FcmToken             []FcmToken            `json:"-"`
	PersonalUserBooks    []PersonalUserBook    `json:"personalUserBooks"`
}

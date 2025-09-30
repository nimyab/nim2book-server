package domain

type User struct {
	Id                   Id                    `json:"id" db:"id"`
	IsAdmin              bool                  `json:"isAdmin" db:"is_admin"`
	IsVIP                bool                  `json:"isVIP" db:"is_vip"`
	GoogleAccount        *GoogleAccount        `json:"googleAccount,omitempty"`
	EmailPasswordAccount *EmailPasswordAccount `json:"emailPasswordAccount,omitempty"`
	FcmToken             *FcmToken             `json:"-"`
}

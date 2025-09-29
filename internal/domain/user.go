package domain

type User struct {
	Id                   Id                    `json:"id" db:"id"`
	IsAdmin              bool                  `json:"isAdmin" db:"is_admin"`
	GoogleAccount        *GoogleAccount        `json:"googleAccount,omitempty"`
	EmailPasswordAccount *EmailPasswordAccount `json:"emailPasswordAccount,omitempty"`
}

package domain

type EmailPasswordAccount struct {
	Id           Id     `json:"id" db:"id"`
	Email        string `json:"email" db:"email"`
	PasswordHash string `json:"-" db:"password_hash"`
}

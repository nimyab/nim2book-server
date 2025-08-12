package domain

type User struct {
	Id           Id     `json:"id" db:"id"`
	Email        string `json:"email" db:"email"`
	PasswordHash string `json:"-" db:"password_hash"`
	IsAdmin      bool   `json:"isAdmin" db:"is_admin"`
}

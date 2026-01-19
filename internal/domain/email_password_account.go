package domain

type EmailPasswordAccount struct {
	Id           Id     `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}

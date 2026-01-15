package google_login

import "github.com/nimyab/nim2book-back/internal/domain"

type Input struct {
	IdToken string `json:"idToken" validate:"required"`
}

type Output struct {
	User         *domain.User `json:"user"`
	AccessToken  string       `json:"accessToken"`
	RefreshToken string       `json:"refreshToken"`
}

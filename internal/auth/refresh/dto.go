package refresh

type Input struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type Output struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

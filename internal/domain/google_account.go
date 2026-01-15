package domain

type GoogleAccount struct {
	Sub           string  `json:"sub"`
	Email         string  `json:"email"`
	EmailVerified bool    `json:"emailVerified"`
	Name          string  `json:"name"`
	Picture       *string `json:"picture"`
}

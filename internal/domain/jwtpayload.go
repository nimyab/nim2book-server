package domain

type JwtPayload struct {
	ID      ID   `json:"id"`
	IsAdmin bool `json:"isAdmin"`
	IsVIP   bool `json:"isVIP"`
}

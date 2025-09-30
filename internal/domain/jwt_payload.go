package domain

type JwtPayload struct {
	Id      Id   `json:"id"`
	IsAdmin bool `json:"isAdmin"`
	IsVIP   bool `json:"isVIP"`
}

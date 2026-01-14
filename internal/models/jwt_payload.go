package models

type JwtPayload struct {
	Id      ID   `json:"id"`
	IsAdmin bool `json:"isAdmin"`
	IsVIP   bool `json:"isVIP"`
}

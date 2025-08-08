package middleware

import (
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	customjwt "github.com/nimyab/nim2book-back/pkg/jwt"
)

func JWT(secretKey string) echo.MiddlewareFunc {
	return echojwt.WithConfig(echojwt.Config{
		SigningKey: secretKey,
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return &customjwt.CustomClaims{}
		},
	})
}

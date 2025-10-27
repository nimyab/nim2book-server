package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/internal/domain"
)

type CustomClaims struct {
	jwt.RegisteredClaims
	domain.JwtPayload
}

func GetUserPayload(c echo.Context) domain.JwtPayload {
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*CustomClaims)
	return claims.JwtPayload
}

func ParseToken(token, secret string) (domain.JwtPayload, error) {
	claims := &CustomClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return domain.JwtPayload{}, err
	}

	return claims.JwtPayload, nil
}

func GenerateTokens(
	payload domain.JwtPayload,
	secret string,
	accessTime,
	refreshTime time.Duration,
) (accessToken string, refreshToken string, err error) {
	accessClaims := CustomClaims{
		JwtPayload: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTime)),
		},
	}

	refreshClaims := CustomClaims{
		JwtPayload: payload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTime)),
		},
	}

	accessToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(secret))
	if err != nil {
		return
	}

	refreshToken, err = jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(secret))
	if err != nil {
		return
	}

	return
}

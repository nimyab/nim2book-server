package jwt

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestJWT(t *testing.T) {
	payload := domain.JwtPayload{
		Id:      uuid.New(),
		IsAdmin: true,
		IsVIP:   false,
	}
	secret := "mysecretkey"

	accessToken, refreshToken, err := GenerateTokens(payload, secret, 15*time.Minute, 7*24*time.Hour)
	assert.Nil(t, err, "GenerateTokens() should not return nil error")

	parsedAccessPayload, err := ParseToken(accessToken, secret)
	assert.Nil(t, err, "ParseToken() should not return nil error")

	assert.Equal(t, parsedAccessPayload, payload)

	parsedRefreshPayload, err := ParseToken(refreshToken, secret)
	assert.Nil(t, err, "ParseToken() should not return nil error")

	assert.Equal(t, parsedRefreshPayload, payload)
}

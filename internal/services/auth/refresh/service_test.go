package refresh

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/jwt"
	"github.com/stretchr/testify/assert"
)

func TestRefresh(t *testing.T) {
	secret := "test-secret"
	accessTime := time.Hour
	refreshTime := time.Hour * 24

	service := New(secret, accessTime, refreshTime)

	// Create a valid refresh token
	payload := domain.JwtPayload{
		ID:      uuid.New(),
		IsAdmin: false,
		IsVIP:   true,
	}
	_, validRefreshToken, _ := jwt.GenerateTokens(payload, secret, accessTime, refreshTime)

	tests := []struct {
		name           string
		inputToken     string
		expectedOutput *Output
		expectedError  error
	}{
		{
			name:          "Success",
			inputToken:    validRefreshToken,
			expectedError: nil,
		},
		{
			name:          "Invalid Token",
			inputToken:    "invalid-token",
			expectedError: ErrParseTokenFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &Input{RefreshToken: tt.inputToken}
			output, err := service.Refresh(input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.NotEmpty(t, output.AccessToken)
				assert.NotEmpty(t, output.RefreshToken)
			}
		})
	}
}

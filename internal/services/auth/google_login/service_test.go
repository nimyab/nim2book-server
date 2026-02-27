package google_login

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/api/idtoken"
)

func TestGoogleLogin(t *testing.T) {
	googleClientID := "test-client-id"
	secret := "test-secret"
	accessTime := time.Hour
	refreshTime := time.Hour * 24

	validPayload := &idtoken.Payload{
		Claims: map[string]any{
			"email":          "test@example.com",
			"sub":            "123456789",
			"email_verified": true,
			"name":           "Test User",
			"picture":        "http://example.com/pic.jpg",
		},
	}

	tests := []struct {
		name           string
		inputToken     string
		mockValidate   func(*MockTokenValidator)
		mockRepo       func(*MockUserRepository)
		expectedOutput *Output
		expectedError  error
	}{
		{
			name:       "Success - Existing User",
			inputToken: "valid-token",
			mockValidate: func(m *MockTokenValidator) {
				m.On("Validate", mock.Anything, "valid-token", googleClientID).Return(validPayload, nil)
			},
			mockRepo: func(m *MockUserRepository) {
				existingAccount := &domain.GoogleAccount{
					Sub:   "123456789",
					Email: "test@example.com",
					User: &domain.User{
						ID: uuid.New(),
					},
				}
				m.On("GetGoogleAccountBySub", mock.Anything, "123456789").Return(existingAccount, nil)
			},
			expectedError: nil,
		},
		{
			name:       "Success - New User",
			inputToken: "valid-token",
			mockValidate: func(m *MockTokenValidator) {
				m.On("Validate", mock.Anything, "valid-token", googleClientID).Return(validPayload, nil)
			},
			mockRepo: func(m *MockUserRepository) {
				m.On("GetGoogleAccountBySub", mock.Anything, "123456789").Return(nil, nil)
				newUser := &domain.User{
					ID: uuid.New(),
				}
				m.On("CreateWithGoogleAccount", mock.Anything, mock.AnythingOfType("*domain.User"), mock.AnythingOfType("*domain.GoogleAccount")).Return(newUser, nil)
			},
			expectedError: nil,
		},
		{
			name:       "Invalid Token",
			inputToken: "invalid-token",
			mockValidate: func(m *MockTokenValidator) {
				m.On("Validate", mock.Anything, "invalid-token", googleClientID).Return(nil, errors.New("invalid token"))
			},
			mockRepo:      func(m *MockUserRepository) {},
			expectedError: ErrInvalidToken,
		},
		{
			name:       "Missing Claims",
			inputToken: "valid-token-missing-claims",
			mockValidate: func(m *MockTokenValidator) {
				payload := &idtoken.Payload{
					Claims: map[string]any{
						"sub": "123456789",
						// Missing email and others
					},
				}
				m.On("Validate", mock.Anything, "valid-token-missing-claims", googleClientID).Return(payload, nil)
			},
			mockRepo:      func(m *MockUserRepository) {},
			expectedError: ErrInvalidGoogleData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockUserRepository(t)
			mockValidator := NewMockTokenValidator(t)

			if tt.mockRepo != nil {
				tt.mockRepo(mockRepo)
			}
			if tt.mockValidate != nil {
				tt.mockValidate(mockValidator)
			}

			service := &Service{
				userRepo:       mockRepo,
				tokenValidator: mockValidator,
				secret:         secret,
				googleClientId: googleClientID,
				accessTime:     accessTime,
				refreshTime:    refreshTime,
			}

			input := &Input{IdToken: tt.inputToken}
			output, err := service.GoogleLogin(context.Background(), input)

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

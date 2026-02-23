package login

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestLogin(t *testing.T) {
	secret := "test-secret"
	accessTime := time.Hour
	refreshTime := time.Hour * 24

	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	tests := []struct {
		name           string
		inputEmail     string
		inputPassword  string
		mockRepo       func(*MockUserRepository)
		expectedOutput *Output
		expectedError  error
	}{
		{
			name:          "Success",
			inputEmail:    "test@example.com",
			inputPassword: password,
			mockRepo: func(m *MockUserRepository) {
				user := &domain.User{
					ID: uuid.New(),
					BasicAccount: &domain.BasicAccount{
						PasswordHash: string(hashedPassword),
					},
				}
				m.On("GetByBasicAccountEmail", mock.Anything, "test@example.com").Return(user, nil)
			},
			expectedError: nil,
		},
		{
			name:          "User Not Found",
			inputEmail:    "notfound@example.com",
			inputPassword: password,
			mockRepo: func(m *MockUserRepository) {
				m.On("GetByBasicAccountEmail", mock.Anything, "notfound@example.com").Return(nil, repository.ErrNotFound)
			},
			expectedError: ErrUserNotFound,
		},
		{
			name:          "User Has No Basic Account",
			inputEmail:    "nobasic@example.com",
			inputPassword: password,
			mockRepo: func(m *MockUserRepository) {
				user := &domain.User{
					ID:           uuid.New(),
					BasicAccount: nil,
				}
				m.On("GetByBasicAccountEmail", mock.Anything, "nobasic@example.com").Return(user, nil)
			},
			expectedError: ErrUserNotFound,
		},
		{
			name:          "Wrong Password",
			inputEmail:    "test@example.com",
			inputPassword: "wrongpassword",
			mockRepo: func(m *MockUserRepository) {
				user := &domain.User{
					ID: uuid.New(),
					BasicAccount: &domain.BasicAccount{
						PasswordHash: string(hashedPassword),
					},
				}
				m.On("GetByBasicAccountEmail", mock.Anything, "test@example.com").Return(user, nil)
			},
			expectedError: ErrPasswordDoNotMatch,
		},
		{
			name:          "Repository Error",
			inputEmail:    "error@example.com",
			inputPassword: password,
			mockRepo: func(m *MockUserRepository) {
				m.On("GetByBasicAccountEmail", mock.Anything, "error@example.com").Return(nil, errors.New("db error"))
			},
			expectedError: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockUserRepository(t)
			if tt.mockRepo != nil {
				tt.mockRepo(mockRepo)
			}

			service := New(mockRepo, secret, accessTime, refreshTime)

			input := &Input{
				Email:    tt.inputEmail,
				Password: tt.inputPassword,
			}

			output, err := service.Login(context.Background(), input)

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

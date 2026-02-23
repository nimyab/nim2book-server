package register

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegister(t *testing.T) {
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
			inputPassword: "password123",
			mockRepo: func(m *MockUserRepository) {
				m.On("GetBasicAccountByEmail", mock.Anything, "test@example.com").Return(nil, repository.ErrNotFound)
				
				newUser := &domain.User{ID: uuid.New()}
				m.On("CreateWithBasicAccount", 
					mock.Anything, 
					mock.MatchedBy(func(u *domain.User) bool { return u != nil }), 
					mock.MatchedBy(func(b *domain.BasicAccount) bool { 
						return b.Email == "test@example.com" 
					}),
				).Return(newUser, nil)
			},
			expectedError: nil,
		},
		{
			name:          "User Already Exists",
			inputEmail:    "existing@example.com",
			inputPassword: "password123",
			mockRepo: func(m *MockUserRepository) {
				existingAccount := &domain.BasicAccount{Email: "existing@example.com"}
				m.On("GetBasicAccountByEmail", mock.Anything, "existing@example.com").Return(existingAccount, nil)
			},
			expectedError: ErrUserAlreadyExist,
		},
		{
			name:          "Repository Error - Get",
			inputEmail:    "error@example.com",
			inputPassword: "password123",
			mockRepo: func(m *MockUserRepository) {
				m.On("GetBasicAccountByEmail", mock.Anything, "error@example.com").Return(nil, errors.New("db error"))
			},
			expectedError: ErrInternal,
		},
		{
			name:          "Repository Error - Create",
			inputEmail:    "test@example.com",
			inputPassword: "password123",
			mockRepo: func(m *MockUserRepository) {
				m.On("GetBasicAccountByEmail", mock.Anything, "test@example.com").Return(nil, repository.ErrNotFound)
				
				m.On("CreateWithBasicAccount", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
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

			service := New(mockRepo)

			input := &Input{
				Email:    tt.inputEmail,
				Password: tt.inputPassword,
			}

			output, err := service.Register(context.Background(), input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.True(t, output.Success)
			}
		})
	}
}

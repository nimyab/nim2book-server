package me

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

func TestMe(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name          string
		input         *Input
		mockRepo      func(*MockUserRepository)
		expectedError error
		expectedUser  *domain.User
	}{
		{
			name:  "Success",
			input: &Input{UserId: userID},
			mockRepo: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(&domain.User{
					ID: userID,
				}, nil)
			},
			expectedError: nil,
			expectedUser: &domain.User{
				ID: userID,
			},
		},
		{
			name:  "User Not Found (Error)",
			input: &Input{UserId: userID},
			mockRepo: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(nil, repository.ErrNotFound)
			},
			expectedError: ErrUserNotFound,
			expectedUser:  nil,
		},
		{
			name:  "User Not Found (Nil)",
			input: &Input{UserId: userID},
			mockRepo: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(nil, nil)
			},
			expectedError: ErrUserNotFound,
			expectedUser:  nil,
		},
		{
			name:  "Repo Error",
			input: &Input{UserId: userID},
			mockRepo: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(nil, errors.New("db error"))
			},
			expectedError: ErrInternal,
			expectedUser:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockUserRepository(t)
			if tt.mockRepo != nil {
				tt.mockRepo(mockRepo)
			}

			service := New(mockRepo)
			output, err := service.Me(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUser, output.User)
			}
		})
	}
}

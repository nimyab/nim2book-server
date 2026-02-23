package metadata

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

func TestUpdateMetadata(t *testing.T) {
	userID := uuid.New()
	newMetadata := domain.JsonB{"key": "value"}

	tests := []struct {
		name          string
		input         *Input
		mockRepo      func(*MockUserRepository)
		expectedError error
		expectedUser  *domain.User
	}{
		{
			name: "Success",
			input: &Input{
				UserId:   userID,
				Metadata: newMetadata,
			},
			mockRepo: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(&domain.User{
					ID: userID,
				}, nil)
				m.On("Update", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
					return u.ID == userID && u.Metadata["key"] == "value"
				})).Return(&domain.User{
					ID:       userID,
					Metadata: newMetadata,
				}, nil)
			},
			expectedError: nil,
			expectedUser: &domain.User{
				ID:       userID,
				Metadata: newMetadata,
			},
		},
		{
			name: "User Not Found",
			input: &Input{
				UserId:   userID,
				Metadata: newMetadata,
			},
			mockRepo: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(nil, repository.ErrNotFound)
			},
			expectedError: ErrUserNotFound,
		},
		{
			name: "Repo Error Get",
			input: &Input{
				UserId:   userID,
				Metadata: newMetadata,
			},
			mockRepo: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(nil, errors.New("db error"))
			},
			expectedError: ErrInternal,
		},
		{
			name: "Repo Error Update",
			input: &Input{
				UserId:   userID,
				Metadata: newMetadata,
			},
			mockRepo: func(m *MockUserRepository) {
				m.On("GetByID", mock.Anything, userID).Return(&domain.User{
					ID: userID,
				}, nil)
				m.On("Update", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
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
			output, err := service.UpdateMetadata(context.Background(), tt.input)

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

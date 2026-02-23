package add_fcm_token

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

func TestAddFcmToken(t *testing.T) {
	userID := uuid.New()
	tests := []struct {
		name           string
		inputToken     string
		inputUserID    domain.ID
		mockRepo       func(*MockFcmTokenRepository)
		expectedOutput *Output
		expectedError  error
	}{
		{
			name:        "Token Already Exists",
			inputToken:  "existing_token",
			inputUserID: userID,
			mockRepo: func(m *MockFcmTokenRepository) {
				m.On("GetByToken", mock.Anything, "existing_token").Return(&domain.FcmToken{Token: "existing_token"}, nil)
			},
			expectedOutput: nil,
			expectedError:  ErrTokenAlreadyAdd,
		},
		{
			name:        "GetByToken Error",
			inputToken:  "token",
			inputUserID: userID,
			mockRepo: func(m *MockFcmTokenRepository) {
				m.On("GetByToken", mock.Anything, "token").Return(nil, errors.New("db error"))
			},
			expectedOutput: nil,
			expectedError:  ErrInternal,
		},
		{
			name:        "Success",
			inputToken:  "new_token",
			inputUserID: userID,
			mockRepo: func(m *MockFcmTokenRepository) {
				m.On("GetByToken", mock.Anything, "new_token").Return(nil, repository.ErrNotFound)
				m.On("Create", mock.Anything, mock.MatchedBy(func(t *domain.FcmToken) bool {
					return t.Token == "new_token" && t.User.ID == userID
				})).Return(&domain.FcmToken{}, nil)
			},
			expectedOutput: &Output{Success: true},
			expectedError:  nil,
		},
		{
			name:        "Create Error",
			inputToken:  "new_token",
			inputUserID: userID,
			mockRepo: func(m *MockFcmTokenRepository) {
				m.On("GetByToken", mock.Anything, "new_token").Return(nil, repository.ErrNotFound)
				m.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("create error"))
			},
			expectedOutput: nil,
			expectedError:  ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockFcmTokenRepository(t)
			if tt.mockRepo != nil {
				tt.mockRepo(mockRepo)
			}

			service := New(mockRepo)

			input := &Input{FcmToken: tt.inputToken}
			output, err := service.AddFcmToken(context.Background(), input, tt.inputUserID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, tt.expectedOutput.Success, output.Success)
			}
		})
	}
}

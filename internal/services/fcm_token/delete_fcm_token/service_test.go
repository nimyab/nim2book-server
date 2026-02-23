package delete_fcm_token

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeleteFcmToken(t *testing.T) {
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
			name:        "Success",
			inputToken:  "token_to_delete",
			inputUserID: userID,
			mockRepo: func(m *MockFcmTokenRepository) {
				m.On("DeleteByToken", mock.Anything, "token_to_delete").Return(nil)
			},
			expectedOutput: &Output{Success: true},
			expectedError:  nil,
		},
		{
			name:        "Delete Error",
			inputToken:  "token_to_delete",
			inputUserID: userID,
			mockRepo: func(m *MockFcmTokenRepository) {
				m.On("DeleteByToken", mock.Anything, "token_to_delete").Return(errors.New("db error"))
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
			output, err := service.DeleteFcmToken(context.Background(), input, tt.inputUserID)

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

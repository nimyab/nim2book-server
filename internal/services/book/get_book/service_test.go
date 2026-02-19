package get_book

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetBook(t *testing.T) {
	tests := []struct {
		name          string
		inputID       domain.ID
		mockReturn    *domain.Book
		mockError     error
		expectedBook  *domain.Book
		expectedError bool
	}{
		{
			name:    "Success",
			inputID: uuid.New(),
			mockReturn: &domain.Book{
				ID:    uuid.New(),
				Title: "Test Book",
			},
			mockError:     nil,
			expectedBook:  &domain.Book{Title: "Test Book"}, // Check specific fields
			expectedError: false,
		},
		{
			name:          "NotFound",
			inputID:       uuid.New(),
			mockReturn:    nil,
			mockError:     errors.New("book not found"),
			expectedBook:  nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := NewMockBookRepository(t)
			service := New(mockRepo)

			// Expectation
			mockRepo.On("GetByID", mock.Anything, tt.inputID).Return(tt.mockReturn, tt.mockError)

			// Execute
			input := &Input{Id: tt.inputID}
			output, err := service.GetBook(context.Background(), input)

			// Assert
			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, tt.mockReturn, output.Book)
			}
		})
	}
}

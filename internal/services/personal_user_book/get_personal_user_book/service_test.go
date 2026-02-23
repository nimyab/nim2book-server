package get_personal_user_book

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

func TestGetPersonalUserBook(t *testing.T) {
	userID := uuid.New()
	bookID := uuid.New()

	tests := []struct {
		name          string
		input         *Input
		mockSetup     func(*MockPersonalBookRepository)
		expectedError error
		expectedBook  *domain.PersonalBook
	}{
		{
			name: "Success",
			input: &Input{
				BookId: bookID,
				UserId: userID,
			},
			mockSetup: func(m *MockPersonalBookRepository) {
				m.On("GetByID", mock.Anything, bookID).Return(&domain.PersonalBook{
					ID:   bookID,
					User: &domain.User{ID: userID},
				}, nil)
			},
			expectedError: nil,
			expectedBook: &domain.PersonalBook{
				ID:   bookID,
				User: &domain.User{ID: userID},
			},
		},
		{
			name: "Book Not Found",
			input: &Input{
				BookId: bookID,
				UserId: userID,
			},
			mockSetup: func(m *MockPersonalBookRepository) {
				m.On("GetByID", mock.Anything, bookID).Return(nil, repository.ErrNotFound)
			},
			expectedError: ErrBookNotFound,
			expectedBook:  nil,
		},
		{
			name: "Repository Error",
			input: &Input{
				BookId: bookID,
				UserId: userID,
			},
			mockSetup: func(m *MockPersonalBookRepository) {
				m.On("GetByID", mock.Anything, bookID).Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
			expectedBook:  nil,
		},
		{
			name: "Forbidden Access",
			input: &Input{
				BookId: bookID,
				UserId: userID,
			},
			mockSetup: func(m *MockPersonalBookRepository) {
				otherUserID := uuid.New()
				m.On("GetByID", mock.Anything, bookID).Return(&domain.PersonalBook{
					ID:   bookID,
					User: &domain.User{ID: otherUserID},
				}, nil)
			},
			expectedError: ErrForbidden,
			expectedBook:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockPersonalBookRepository(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := New(mockRepo)
			output, err := service.GetPersonalUserBook(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.expectedError.Error() == "db error" {
					assert.Equal(t, tt.expectedError.Error(), err.Error())
				} else {
					assert.Equal(t, tt.expectedError, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBook, output.Book)
			}
		})
	}
}

package get_personal_user_books

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

func TestGetPersonalUserBooks(t *testing.T) {
	userID := uuid.New()
	bookID := uuid.New()
	genreID := uuid.New()

	tests := []struct {
		name          string
		input         *Input
		mockSetup     func(*MockPersonalBookRepository)
		expectedError error
		expectedBooks []domain.PersonalBook
	}{
		{
			name: "Success",
			input: &Input{
				UserId:  userID,
				Page:    1,
				Title:   "Book Title",
				Author:  "Author Name",
				GenreId: &genreID,
			},
			mockSetup: func(m *MockPersonalBookRepository) {
				m.On("SearchByUserWithFilters", mock.Anything, userID, "Book Title", "Author Name", &genreID, repository.QueryOptions{
					Limit:  10,
					Offset: 0,
				}).Return([]*domain.PersonalBook{
					{ID: bookID, Title: "Book Title"},
				}, nil)
			},
			expectedError: nil,
			expectedBooks: []domain.PersonalBook{
				{ID: bookID, Title: "Book Title"},
			},
		},
		{
			name: "Repository Error",
			input: &Input{
				UserId: userID,
				Page:   1,
			},
			mockSetup: func(m *MockPersonalBookRepository) {
				m.On("SearchByUserWithFilters", mock.Anything, userID, "", "", (*domain.ID)(nil), repository.QueryOptions{
					Limit:  10,
					Offset: 0,
				}).Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("db error"),
			expectedBooks: nil,
		},
		{
			name: "Pagination",
			input: &Input{
				UserId: userID,
				Page:   2,
			},
			mockSetup: func(m *MockPersonalBookRepository) {
				m.On("SearchByUserWithFilters", mock.Anything, userID, "", "", (*domain.ID)(nil), repository.QueryOptions{
					Limit:  10,
					Offset: 10,
				}).Return([]*domain.PersonalBook{}, nil)
			},
			expectedError: nil,
			expectedBooks: []domain.PersonalBook{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockPersonalBookRepository(t)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			service := New(mockRepo)
			output, err := service.GetPersonalUserBooks(context.Background(), tt.input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBooks, output.Books)
			}
		})
	}
}

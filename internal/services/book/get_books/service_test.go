package get_books

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

func TestGetBooks(t *testing.T) {
	tests := []struct {
		name           string
		inputPage      int
		inputTitle     string
		inputAuthor    string
		inputGenreID   *domain.ID
		mockRepo       func(*MockBookRepository)
		expectedOutput *Output
		expectedError  error
	}{
		{
			name:      "Success - No Filters",
			inputPage: 1,
			mockRepo: func(m *MockBookRepository) {
				books := []*domain.Book{
					{Title: "Book 1"},
					{Title: "Book 2"},
				}
				m.On("SearchWithFilters", mock.Anything, "", "", (*domain.ID)(nil), repository.QueryOptions{Limit: 10, Offset: 0}).Return(books, nil)
			},
			expectedOutput: &Output{
				Books: []domain.Book{{Title: "Book 1"}, {Title: "Book 2"}},
			},
			expectedError: nil,
		},
		{
			name:        "Success - With Filters",
			inputPage:   2,
			inputTitle:  "Title",
			inputAuthor: "Author",
			inputGenreID: func() *domain.ID {
				id := uuid.New()
				return &id
			}(),
			mockRepo: func(m *MockBookRepository) {
				books := []*domain.Book{{Title: "Filtered Book"}}
				m.On("SearchWithFilters",
					mock.Anything,
					"Title",
					"Author",
					mock.AnythingOfType("*uuid.UUID"),
					repository.QueryOptions{Limit: 10, Offset: 10},
				).Return(books, nil)
			},
			expectedOutput: &Output{
				Books: []domain.Book{{Title: "Filtered Book"}},
			},
			expectedError: nil,
		},
		{
			name:      "Repository Error",
			inputPage: 1,
			mockRepo: func(m *MockBookRepository) {
				m.On("SearchWithFilters", mock.Anything, "", "", (*domain.ID)(nil), repository.QueryOptions{Limit: 10, Offset: 0}).Return(nil, errors.New("db error"))
			},
			expectedOutput: nil,
			expectedError:  errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockBookRepository(t)
			if tt.mockRepo != nil {
				tt.mockRepo(mockRepo)
			}

			service := New(mockRepo)

			input := &Input{
				Page:    tt.inputPage,
				Title:   tt.inputTitle,
				Author:  tt.inputAuthor,
				GenreId: tt.inputGenreID,
			}

			output, err := service.GetBooks(context.Background(), input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, tt.expectedOutput.Books, output.Books)
			}
		})
	}
}

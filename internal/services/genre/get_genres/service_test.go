package get_genres

import (
	"context"
	"errors"
	"testing"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetGenres(t *testing.T) {
	tests := []struct {
		name           string
		mockRepo       func(*MockGenreRepository)
		expectedOutput *Output
		expectedError  error
	}{
		{
			name: "Success",
			mockRepo: func(m *MockGenreRepository) {
				m.On("List", mock.Anything, repository.QueryOptions{}).Return([]*domain.Genre{
					{Name: "Fantasy"},
					{Name: "Sci-Fi"},
				}, nil)
			},
			expectedOutput: &Output{
				Genres: []domain.Genre{
					{Name: "Fantasy"},
					{Name: "Sci-Fi"},
				},
			},
			expectedError: nil,
		},
		{
			name: "Error",
			mockRepo: func(m *MockGenreRepository) {
				m.On("List", mock.Anything, repository.QueryOptions{}).Return(nil, errors.New("db error"))
			},
			expectedOutput: nil,
			expectedError:  errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockGenreRepository(t)
			if tt.mockRepo != nil {
				tt.mockRepo(mockRepo)
			}

			service := New(mockRepo)

			output, err := service.GetGenres(context.Background())

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, len(tt.expectedOutput.Genres), len(output.Genres))
				if len(output.Genres) > 0 {
					assert.Equal(t, tt.expectedOutput.Genres[0].Name, output.Genres[0].Name)
				}
			}
		})
	}
}

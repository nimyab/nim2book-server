package get_genre

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetGenre(t *testing.T) {
	genreID := uuid.New()
	tests := []struct {
		name           string
		inputID        domain.ID
		mockRepo       func(*MockGenreRepository)
		expectedOutput *Output
		expectedError  error
	}{
		{
			name:    "Success",
			inputID: genreID,
			mockRepo: func(m *MockGenreRepository) {
				m.On("GetByID", mock.Anything, genreID).Return(&domain.Genre{ID: genreID, Name: "Fantasy"}, nil)
			},
			expectedOutput: &Output{
				Genre: &domain.Genre{ID: genreID, Name: "Fantasy"},
			},
			expectedError: nil,
		},
		{
			name:    "Error",
			inputID: genreID,
			mockRepo: func(m *MockGenreRepository) {
				m.On("GetByID", mock.Anything, genreID).Return(nil, errors.New("db error"))
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

			input := &Input{Id: tt.inputID}
			output, err := service.GetGenre(context.Background(), input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, tt.expectedOutput.Genre.ID, output.Genre.ID)
				assert.Equal(t, tt.expectedOutput.Genre.Name, output.Genre.Name)
			}
		})
	}
}

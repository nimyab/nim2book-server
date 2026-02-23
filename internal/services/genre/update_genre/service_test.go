package update_genre

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

func TestUpdateGenre(t *testing.T) {
	genreID := uuid.New()
	tests := []struct {
		name           string
		inputID        domain.ID
		inputName      string
		mockRepo       func(*MockGenreRepository)
		expectedOutput *Output
		expectedError  error
	}{
		{
			name:      "Success",
			inputID:   genreID,
			inputName: "Updated Name",
			mockRepo: func(m *MockGenreRepository) {
				m.On("Update", mock.Anything, mock.MatchedBy(func(g *domain.Genre) bool {
					return g.ID == genreID && g.Name == "Updated Name"
				})).Return(&domain.Genre{ID: genreID, Name: "Updated Name"}, nil)
			},
			expectedOutput: &Output{
				Genre: &domain.Genre{ID: genreID, Name: "Updated Name"},
			},
			expectedError: nil,
		},
		{
			name:      "Not Found",
			inputID:   genreID,
			inputName: "Updated Name",
			mockRepo: func(m *MockGenreRepository) {
				m.On("Update", mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)
			},
			expectedOutput: nil,
			expectedError:  ErrGenreNotFound,
		},
		{
			name:      "Duplicate Key",
			inputID:   genreID,
			inputName: "Existing Name",
			mockRepo: func(m *MockGenreRepository) {
				m.On("Update", mock.Anything, mock.Anything).Return(nil, repository.ErrDuplicateKey)
			},
			expectedOutput: nil,
			expectedError:  ErrGenreAlreadyExists,
		},
		{
			name:      "Internal Error",
			inputID:   genreID,
			inputName: "Updated Name",
			mockRepo: func(m *MockGenreRepository) {
				m.On("Update", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			expectedOutput: nil,
			expectedError:  ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockGenreRepository(t)
			if tt.mockRepo != nil {
				tt.mockRepo(mockRepo)
			}

			service := New(mockRepo)

			input := &Input{Id: tt.inputID, Name: tt.inputName}
			output, err := service.UpdateGenre(context.Background(), input)

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

package create_genre

import (
	"context"
	"errors"
	"testing"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateGenre(t *testing.T) {
	tests := []struct {
		name           string
		inputName      string
		mockRepo       func(*MockGenreRepository)
		expectedOutput *Output
		expectedError  error
	}{
		{
			name:      "Success",
			inputName: "Fantasy",
			mockRepo: func(m *MockGenreRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(g *domain.Genre) bool {
					return g.Name == "Fantasy"
				})).Return(&domain.Genre{Name: "Fantasy"}, nil)
			},
			expectedOutput: &Output{
				Genre: &domain.Genre{Name: "Fantasy"},
			},
			expectedError: nil,
		},
		{
			name:      "Duplicate Key",
			inputName: "Fantasy",
			mockRepo: func(m *MockGenreRepository) {
				m.On("Create", mock.Anything, mock.Anything).Return(nil, repository.ErrDuplicateKey)
			},
			expectedOutput: nil,
			expectedError:  ErrGenreAlreadyExists,
		},
		{
			name:      "Internal Error",
			inputName: "Fantasy",
			mockRepo: func(m *MockGenreRepository) {
				m.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
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

			input := &Input{Name: tt.inputName}
			output, err := service.CreateGenre(context.Background(), input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, tt.expectedOutput.Genre.Name, output.Genre.Name)
			}
		})
	}
}

package update_book

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateBook(t *testing.T) {
	tests := []struct {
		name           string
		inputID        domain.ID
		inputTitle     *string
		inputAuthor    *string
		mockBookRepo   func(*MockBookRepository)
		mockAuthorRepo func(*MockAuthorRepository)
		mockS3         func(*MockS3)
		cover          *multipart.FileHeader
		expectedOutput *Output
		expectedError  error
	}{
		{
			name:    "Success - Update Title Only",
			inputID: uuid.New(),
			inputTitle: func() *string {
				s := "New Title"
				return &s
			}(),
			mockBookRepo: func(m *MockBookRepository) {
				book := &domain.Book{ID: uuid.New(), Title: "Old Title"}
				m.On("GetByID", mock.Anything, mock.Anything).Return(book, nil)
				m.On("Update", mock.Anything, mock.Anything).Return(book, nil)
			},
			mockAuthorRepo: func(m *MockAuthorRepository) {},
			mockS3:         func(m *MockS3) {},
			expectedOutput: &Output{
				Book: &domain.Book{Title: "Old Title"}, // Note: Update returns the book pointer passed or new? Mock returns 'book'.
			},
			expectedError: nil,
		},
		{
			name:    "Success - Update Author Only",
			inputID: uuid.New(),
			inputAuthor: func() *string {
				s := "New Author"
				return &s
			}(),
			mockBookRepo: func(m *MockBookRepository) {
				book := &domain.Book{ID: uuid.New(), Title: "Title"}
				m.On("GetByID", mock.Anything, mock.Anything).Return(book, nil)
				m.On("Update", mock.Anything, mock.Anything).Return(book, nil)
			},
			mockAuthorRepo: func(m *MockAuthorRepository) {
				author := &domain.Author{Name: "New Author"}
				m.On("GetOrCreate", mock.Anything, "New Author").Return(author, nil)
			},
			mockS3: func(m *MockS3) {},
			expectedOutput: &Output{
				Book: &domain.Book{Title: "Title"},
			},
			expectedError: nil,
		},
		{
			name:    "Book Not Found",
			inputID: uuid.New(),
			mockBookRepo: func(m *MockBookRepository) {
				m.On("GetByID", mock.Anything, mock.Anything).Return(nil, repository.ErrNotFound)
			},
			mockAuthorRepo: func(m *MockAuthorRepository) {},
			mockS3:         func(m *MockS3) {},
			expectedError:  ErrBookNotFound,
		},
		{
			name:    "Repository Error",
			inputID: uuid.New(),
			mockBookRepo: func(m *MockBookRepository) {
				m.On("GetByID", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			mockAuthorRepo: func(m *MockAuthorRepository) {},
			mockS3:         func(m *MockS3) {},
			expectedError:  ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBookRepo := NewMockBookRepository(t)
			mockAuthorRepo := NewMockAuthorRepository(t)
			mockS3 := NewMockS3(t)

			if tt.mockBookRepo != nil {
				tt.mockBookRepo(mockBookRepo)
			}
			if tt.mockAuthorRepo != nil {
				tt.mockAuthorRepo(mockAuthorRepo)
			}
			if tt.mockS3 != nil {
				tt.mockS3(mockS3)
			}

			service := New(mockBookRepo, mockAuthorRepo, mockS3)

			input := &Input{
				Id:     tt.inputID,
				Title:  tt.inputTitle,
				Author: tt.inputAuthor,
			}

			output, err := service.UpdateBook(context.Background(), input, tt.cover)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.NotNil(t, output.Book)
			}
		})
	}
}

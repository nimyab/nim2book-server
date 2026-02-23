package update_personal_user_book

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

func TestUpdatePersonalUserBook(t *testing.T) {
	userID := uuid.New()
	bookID := uuid.New()
	authorID := uuid.New()

	newTitle := "New Title"
	newAuthor := "New Author"

	tests := []struct {
		name          string
		input         *Input
		mockRepo      func(*MockPersonalBookRepository)
		mockAuthor    func(*MockAuthorRepository)
		mockS3        func(*MockS3)
		expectedError error
		expectedBook  *domain.PersonalBook
	}{
		{
			name: "Success Update Title And Author",
			input: &Input{
				Id:     bookID,
				UserId: userID,
				Title:  &newTitle,
				Author: &newAuthor,
			},
			mockRepo: func(m *MockPersonalBookRepository) {
				m.On("GetByID", mock.Anything, bookID).Return(&domain.PersonalBook{
					ID:   bookID,
					User: &domain.User{ID: userID},
					Title: "Old Title",
				}, nil)
				m.On("Update", mock.Anything, mock.MatchedBy(func(b *domain.PersonalBook) bool {
					return b.Title == newTitle && b.Author.Name == newAuthor
				})).Return(&domain.PersonalBook{
					ID:     bookID,
					Title:  newTitle,
					Author: &domain.Author{Name: newAuthor},
				}, nil)
			},
			mockAuthor: func(m *MockAuthorRepository) {
				m.On("GetOrCreate", mock.Anything, newAuthor).Return(&domain.Author{
					ID:   authorID,
					Name: newAuthor,
				}, nil)
			},
			mockS3: func(m *MockS3) {},
			expectedError: nil,
			expectedBook: &domain.PersonalBook{
				ID:     bookID,
				Title:  newTitle,
				Author: &domain.Author{Name: newAuthor},
			},
		},
		{
			name: "Book Not Found",
			input: &Input{
				Id:     bookID,
				UserId: userID,
			},
			mockRepo: func(m *MockPersonalBookRepository) {
				m.On("GetByID", mock.Anything, bookID).Return(nil, repository.ErrNotFound)
			},
			mockAuthor: func(m *MockAuthorRepository) {},
			mockS3: func(m *MockS3) {},
			expectedError: ErrBookNotFound,
		},
		{
			name: "Forbidden",
			input: &Input{
				Id:     bookID,
				UserId: userID,
			},
			mockRepo: func(m *MockPersonalBookRepository) {
				otherUserID := uuid.New()
				m.On("GetByID", mock.Anything, bookID).Return(&domain.PersonalBook{
					ID:   bookID,
					User: &domain.User{ID: otherUserID},
				}, nil)
			},
			mockAuthor: func(m *MockAuthorRepository) {},
			mockS3: func(m *MockS3) {},
			expectedError: ErrForbidden,
		},
		{
			name: "Repo Error Get",
			input: &Input{
				Id:     bookID,
				UserId: userID,
			},
			mockRepo: func(m *MockPersonalBookRepository) {
				m.On("GetByID", mock.Anything, bookID).Return(nil, errors.New("db error"))
			},
			mockAuthor: func(m *MockAuthorRepository) {},
			mockS3: func(m *MockS3) {},
			expectedError: ErrInternalServer,
		},
		{
			name: "Repo Error Update",
			input: &Input{
				Id:     bookID,
				UserId: userID,
				Title:  &newTitle,
			},
			mockRepo: func(m *MockPersonalBookRepository) {
				m.On("GetByID", mock.Anything, bookID).Return(&domain.PersonalBook{
					ID:   bookID,
					User: &domain.User{ID: userID},
				}, nil)
				m.On("Update", mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			mockAuthor: func(m *MockAuthorRepository) {},
			mockS3: func(m *MockS3) {},
			expectedError: ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockPersonalBookRepository(t)
			mockAuthor := NewMockAuthorRepository(t)
			mockS3 := NewMockS3(t)

			if tt.mockRepo != nil {
				tt.mockRepo(mockRepo)
			}
			if tt.mockAuthor != nil {
				tt.mockAuthor(mockAuthor)
			}
			if tt.mockS3 != nil {
				tt.mockS3(mockS3)
			}

			service := New(mockRepo, mockAuthor, mockS3)
			output, err := service.UpdatePersonalUserBook(context.Background(), tt.input, nil)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBook, output.Book)
			}
		})
	}
}

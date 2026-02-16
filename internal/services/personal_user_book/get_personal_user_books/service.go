package get_personal_user_books

import (
	"context"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
)

const booksPerPage = 10

type PersonalBookRepository interface {
	SearchByUserWithFilters(ctx context.Context, userID domain.ID, title, authorName string, genreID *domain.ID, opts repository.QueryOptions) ([]*domain.PersonalBook, error)
}

type Service struct {
	personalBookRepo PersonalBookRepository
}

func New(personalBookRepo PersonalBookRepository) *Service {
	return &Service{personalBookRepo: personalBookRepo}
}

func (s *Service) GetPersonalUserBooks(ctx context.Context, input *Input) (*Output, error) {
	const operation = "personal_user_book.get_personal_user_books.GetPersonalUserBooks"

	offset := (input.Page - 1) * booksPerPage

	books, err := s.personalBookRepo.SearchByUserWithFilters(
		ctx,
		input.UserId,
		input.Title,
		input.Author,
		input.GenreId,
		repository.QueryOptions{
			Limit:  booksPerPage,
			Offset: offset,
		},
	)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	// Конвертируем в слайс значений для совместимости с Output
	result := make([]domain.PersonalBook, len(books))
	for i, book := range books {
		result[i] = *book
	}

	return &Output{Books: result}, nil
}

package get_books

import (
	"context"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
)

const booksPerPage = 10

type BookRepository interface {
	SearchWithFilters(ctx context.Context, title, authorName string, genreID *domain.ID, opts repository.QueryOptions) ([]*domain.Book, error)
}

type Service struct {
	bookRepo BookRepository
}

func New(bookRepo BookRepository) *Service {
	return &Service{
		bookRepo: bookRepo,
	}
}

func (s *Service) GetBooks(ctx context.Context, input *Input) (*Output, error) {
	const operation = "book.get_books.GetBooks"

	offset := (input.Page - 1) * booksPerPage

	books, err := s.bookRepo.SearchWithFilters(
		ctx,
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
	result := make([]domain.Book, len(books))
	for i, book := range books {
		result[i] = *book
	}

	return &Output{Books: result}, nil
}

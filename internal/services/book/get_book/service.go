package get_book

import (
	"context"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/helpers"
)

type Postgres interface {
	GetBook(ctx context.Context, id domain.Id) (*domain.Book, error)
	GetBookGenres(ctx context.Context, bookId domain.Id) ([]domain.Genre, error)
}

type Service struct {
	pg Postgres
}

func New(pg Postgres) *Service {
	return &Service{
		pg: pg,
	}
}

func (s *Service) GetBook(input *Input) (*Output, error) {
	const operation = "book.get_book.GetBook"

	book, err := s.pg.GetBook(context.Background(), input.Id)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	// Загружаем жанры книги
	if err := helpers.EnrichBookWithGenres(context.Background(), book, s.pg, operation); err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	output := &Output{Book: book}
	return output, nil
}

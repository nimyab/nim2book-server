package get_books

import (
	"context"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/adapter/postgres_sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/helpers"
)

type Postgres interface {
	GetBooks(ctx context.Context, query postgres_sqlc.GetBooksQuery) ([]domain.Book, error)
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

func (s *Service) GetBooks(input *Input) (*Output, error) {
	const operation = "book.get_books.GetBooks"

	books, err := s.pg.GetBooks(context.Background(), postgres_sqlc.GetBooksQuery{
		Author:  input.Author,
		Title:   input.Title,
		GenreId: input.GenreId,
		Page:    input.Page,
	})
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	// Загружаем жанры для каждой книги
	if err := helpers.EnrichBooksWithGenres(context.Background(), books, s.pg, operation); err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	return &Output{Books: books}, nil
}

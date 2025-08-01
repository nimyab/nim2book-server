package get_books

import (
	"context"
	"github.com/nimyab/nim2book-back/internal/adapter/postgres"
	"github.com/nimyab/nim2book-back/internal/domain"
	"log/slog"
)

type Postgres interface {
	GetBooks(ctx context.Context, query postgres.GetBooksQuery) ([]domain.Book, error)
}

type Service struct {
	pg Postgres
}

var service *Service

func New(pg Postgres) *Service {
	service = &Service{
		pg: pg,
	}
	return service
}

func (s *Service) GetBooks(input *Input) (*Output, error) {
	const operation = "postgres.GetBooks"

	books, err := s.pg.GetBooks(context.Background(), postgres.GetBooksQuery{
		Author: input.Author,
		Title:  input.Title,
		Page:   input.Page,
	})
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	return &Output{Books: books}, nil
}

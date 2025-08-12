package get_book

import (
	"context"
	"github.com/nimyab/nim2book-back/internal/domain"
	"log/slog"
)

type Postgres interface {
	GetBook(ctx context.Context, id domain.Id) (*domain.Book, error)
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

func (s *Service) GetBook(input *Input) (*Output, error) {
	const operation = "postgres.GetBook"

	book, err := s.pg.GetBook(context.Background(), input.Id)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	output := &Output{Book: book}
	return output, nil
}

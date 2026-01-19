package delete_genre

import (
	"context"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
)

var (
	ErrInternalServer = error(nil)
)

type Postgres interface {
	DeleteGenre(ctx context.Context, id domain.Id) error
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

func (s *Service) DeleteGenre(input *Input) (*Output, error) {
	const operation = "genre.delete_genre.DeleteGenre"

	err := s.pg.DeleteGenre(context.Background(), input.Id)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	return &Output{Success: true}, nil
}

package get_genre

import (
	"context"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
)

type Postgres interface {
	GetGenreById(ctx context.Context, id domain.Id) (*domain.Genre, error)
}

type Service struct {
	pg Postgres
}

func New(pg Postgres) *Service {
	return &Service{pg: pg}
}

func (s *Service) GetGenre(input *Input) (*Output, error) {
	const operation = "genre.get_genre.GetGenre"

	genre, err := s.pg.GetGenreById(context.Background(), input.Id)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	return &Output{Genre: genre}, nil
}

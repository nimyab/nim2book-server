package get_genres

import (
	"context"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
)

type Postgres interface {
	GetAllGenres(ctx context.Context) ([]domain.Genre, error)
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

func (s *Service) GetGenres() (*Output, error) {
	const operation = "genre.get_genres.GetGenres"

	genres, err := s.pg.GetAllGenres(context.Background())
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	return &Output{Genres: genres}, nil
}

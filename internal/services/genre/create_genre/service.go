package create_genre

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/adapter/postgres_sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
)

var (
	ErrGenreAlreadyExists = errors.New("genre already exists")
	ErrInternalServer     = errors.New("internal server error")
)

type Postgres interface {
	CreateGenre(ctx context.Context, genre *domain.Genre) (*domain.Genre, error)
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

func (s *Service) CreateGenre(input *Input) (*Output, error) {
	const operation = "genre.create_genre.CreateGenre"

	genre := &domain.Genre{
		Name: input.Name,
	}

	createdGenre, err := s.pg.CreateGenre(context.Background(), genre)
	if errors.Is(err, postgres_sqlc.ErrGenreAlreadyExists) {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrGenreAlreadyExists
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternalServer
	}

	return &Output{Genre: createdGenre}, nil
}

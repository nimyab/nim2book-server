package update_genre

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/adapter/postgres_sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
)

var (
	ErrGenreNotFound      = errors.New("genre not found")
	ErrGenreAlreadyExists = errors.New("genre already exists")
	ErrInternalServer     = errors.New("internal server error")
)

type Postgres interface {
	UpdateGenre(ctx context.Context, genre *domain.Genre) (*domain.Genre, error)
}

type Service struct {
	pg Postgres
}

func New(pg Postgres) *Service {
	return &Service{pg: pg}
}

func (s *Service) UpdateGenre(input *Input) (*Output, error) {
	const operation = "genre.update_genre.UpdateGenre"

	genre := &domain.Genre{
		Id:   input.Id,
		Name: input.Name,
	}

	updatedGenre, err := s.pg.UpdateGenre(context.Background(), genre)
	if errors.Is(err, postgres_sqlc.ErrGenreNotFound) {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrGenreNotFound
	}
	if errors.Is(err, postgres_sqlc.ErrGenreAlreadyExists) {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrGenreAlreadyExists
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternalServer
	}

	return &Output{Genre: updatedGenre}, nil
}

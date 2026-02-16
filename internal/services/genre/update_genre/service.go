package update_genre

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
)

var (
	ErrGenreNotFound      = errors.New("genre not found")
	ErrGenreAlreadyExists = errors.New("genre already exists")
	ErrInternalServer     = errors.New("internal server error")
)

type GenreRepository interface {
	Update(ctx context.Context, genre *domain.Genre) (*domain.Genre, error)
}

type Service struct {
	genreRepo GenreRepository
}

func New(genreRepo GenreRepository) *Service {
	return &Service{genreRepo: genreRepo}
}

func (s *Service) UpdateGenre(ctx context.Context, input *Input) (*Output, error) {
	const operation = "genre.update_genre.UpdateGenre"

	genre := &domain.Genre{
		ID:   input.Id,
		Name: input.Name,
	}

	updatedGenre, err := s.genreRepo.Update(ctx, genre)
	if errors.Is(err, repository.ErrNotFound) {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrGenreNotFound
	}
	if errors.Is(err, repository.ErrDuplicateKey) {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrGenreAlreadyExists
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternalServer
	}

	return &Output{Genre: updatedGenre}, nil
}

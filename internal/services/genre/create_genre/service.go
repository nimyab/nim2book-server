package create_genre

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
)

var (
	ErrGenreAlreadyExists = errors.New("genre already exists")
	ErrInternalServer     = errors.New("internal server error")
)

type GenreRepository interface {
	Create(ctx context.Context, genre *domain.Genre) (*domain.Genre, error)
}

type Service struct {
	genreRepo GenreRepository
}

func New(genreRepo GenreRepository) *Service {
	return &Service{genreRepo: genreRepo}
}

func (s *Service) CreateGenre(input *Input) (*Output, error) {
	const operation = "genre.create_genre.CreateGenre"

	genre := &domain.Genre{
		Name: input.Name,
	}

	createdGenre, err := s.genreRepo.Create(context.Background(), genre)
	if errors.Is(err, repository.ErrDuplicateKey) {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrGenreAlreadyExists
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternalServer
	}

	return &Output{Genre: createdGenre}, nil
}

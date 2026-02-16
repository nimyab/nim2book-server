package get_genre

import (
	"context"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
)

type GenreRepository interface {
	GetByID(ctx context.Context, id domain.ID) (*domain.Genre, error)
}

type Service struct {
	genreRepo GenreRepository
}

func New(genreRepo GenreRepository) *Service {
	return &Service{genreRepo: genreRepo}
}

func (s *Service) GetGenre(ctx context.Context, input *Input) (*Output, error) {
	const operation = "genre.get_genre.GetGenre"

	genre, err := s.genreRepo.GetByID(ctx, input.Id)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	return &Output{Genre: genre}, nil
}

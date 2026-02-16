package delete_genre

import (
	"context"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
)

type GenreRepository interface {
	Delete(ctx context.Context, id domain.ID) error
}

type Service struct {
	genreRepo GenreRepository
}

func New(genreRepo GenreRepository) *Service {
	return &Service{genreRepo: genreRepo}
}

func (s *Service) DeleteGenre(ctx context.Context, input *Input) (*Output, error) {
	const operation = "genre.delete_genre.DeleteGenre"

	err := s.genreRepo.Delete(ctx, input.Id)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	return &Output{Success: true}, nil
}

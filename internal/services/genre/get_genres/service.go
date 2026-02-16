package get_genres

import (
	"context"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
)

type GenreRepository interface {
	List(ctx context.Context, opts repository.QueryOptions) ([]*domain.Genre, error)
}

type Service struct {
	genreRepo GenreRepository
}

func New(genreRepo GenreRepository) *Service {
	return &Service{genreRepo: genreRepo}
}

func (s *Service) GetGenres(ctx context.Context) (*Output, error) {
	const operation = "genre.get_genres.GetGenres"

	// Получить все жанры без пагинации
	genres, err := s.genreRepo.List(ctx, repository.QueryOptions{})
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	// Конвертируем в слайс значений для совместимости с Output
	result := make([]domain.Genre, len(genres))
	for i, genre := range genres {
		result[i] = *genre
	}

	return &Output{Genres: result}, nil
}

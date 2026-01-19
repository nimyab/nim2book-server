package metadata

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/adapter/postgres_sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/helpers"
)

type Postgres interface {
	UpdateMetadata(ctx context.Context, newMetadata map[string]any, userId domain.Id) (*domain.User, error)
	GetPersonalUserBooks(ctx context.Context, query postgres_sqlc.GetPersonalUserBooksQuery) ([]domain.PersonalUserBook, error)
	GetPersonalUserBookGenres(ctx context.Context, bookId domain.Id) ([]domain.Genre, error)
}

type Service struct {
	pg Postgres
}

var service *Service

var (
	ErrInternal     = errors.New("internal error")
	ErrUserNotFound = errors.New("user not found")
)

func New(pg Postgres) *Service {
	service = &Service{
		pg: pg,
	}
	return service
}

func (s *Service) UpdateMetadata(input *Input) (*Output, error) {
	const operation = "user.metadata.UpdateMetadata"

	user, err := s.pg.UpdateMetadata(context.Background(), input.Metadata, input.UserId)
	if errors.Is(err, postgres_sqlc.ErrUserNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	// Загружаем персональные книги пользователя вместе с жанрами
	if err := helpers.EnrichUserWithPersonalBooksAndGenres(context.Background(), user, s.pg, operation); err != nil {
		return nil, ErrInternal
	}

	return &Output{User: user}, nil
}

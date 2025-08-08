package me

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/adapter/postgres"
	"github.com/nimyab/nim2book-back/internal/domain"
	"log/slog"
)

type Postgres interface {
	GetUserById(ctx context.Context, userId uuid.UUID) (*domain.User, error)
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

func (s *Service) Me(input *Input) (*Output, error) {
	const operation = "user.me.Me"

	user, err := s.pg.GetUserById(context.Background(), input.UserId)
	if errors.Is(err, postgres.ErrUserNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	return &Output{User: user}, nil
}

package me

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
)

type UserRepository interface {
	GetByID(ctx context.Context, id domain.ID) (*domain.User, error)
}

type Service struct {
	userRepo UserRepository
}

var (
	ErrInternal     = errors.New("internal error")
	ErrUserNotFound = errors.New("user not found")
)

func New(userRepo UserRepository) *Service {
	return &Service{userRepo: userRepo}
}

func (s *Service) Me(input *Input) (*Output, error) {
	const operation = "user.me.Me"

	user, err := s.userRepo.GetByID(context.Background(), input.UserId)
	if errors.Is(err, repository.ErrNotFound) || user == nil {
		return nil, ErrUserNotFound
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	return &Output{User: user}, nil
}

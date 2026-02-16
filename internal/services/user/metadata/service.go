package metadata

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
)

type UserRepository interface {
	GetByID(ctx context.Context, id domain.ID) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) (*domain.User, error)
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

func (s *Service) UpdateMetadata(ctx context.Context, input *Input) (*Output, error) {
	const operation = "user.metadata.UpdateMetadata"

	// Получаем пользователя
	user, err := s.userRepo.GetByID(ctx, input.UserId)
	if errors.Is(err, repository.ErrNotFound) || user == nil {
		return nil, ErrUserNotFound
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	// Обновляем metadata
	user.Metadata = input.Metadata
	updatedUser, err := s.userRepo.Update(ctx, user)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	return &Output{User: updatedUser}, nil
}

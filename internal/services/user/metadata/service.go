package metadata

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/internal/repositories"
)

type UserRepo interface {
	UpdateMetadata(ctx context.Context, newMetadata map[string]any, userId models.ID) (*models.User, error)
}

type Service struct {
	userRepo *repositories.UserRepository
}

var service *Service

var (
	ErrInternal     = errors.New("internal error")
	ErrUserNotFound = errors.New("user not found")
)

func New(userRepo *repositories.UserRepository) *Service {
	service = &Service{
		userRepo: userRepo,
	}
	return service
}

func (s *Service) UpdateMetadata(input *Input) (*Output, error) {
	const operation = "user.metadata.UpdateMetadata"

	err := s.userRepo.UpdateMetadata(context.Background(), input.UserId, input.Metadata)
	if errors.Is(err, repositories.ErrUserNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	user, err := s.userRepo.GetUserById(context.Background(), input.UserId)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	return &Output{User: user}, nil
}

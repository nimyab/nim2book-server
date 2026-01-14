package add_fcm_token

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/internal/repositories"
)

// FcmTokenRepository defines the interface for fcm token repository operations needed by this service
type FcmTokenRepository interface {
	AddFcmToken(ctx context.Context, token string, userId uuid.UUID) (*models.FcmToken, error)
}

type Service struct {
	fcmTokenRepo FcmTokenRepository
}

var service *Service

var (
	ErrInternal        = errors.New("internal error")
	ErrTokenAlreadyAdd = errors.New("token already add")
)

func New(fcmTokenRepo FcmTokenRepository) *Service {
	service = &Service{
		fcmTokenRepo: fcmTokenRepo,
	}
	return service
}

func (s *Service) AddFcmToken(input *Input, userId models.ID) (*Output, error) {
	const operation = "fcm_token.add_fcm_token.AddFcmToken"

	_, err := s.fcmTokenRepo.AddFcmToken(context.Background(), input.FcmToken, userId)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation), slog.String("token", input.FcmToken))
		if errors.Is(err, repositories.ErrFcmTokenAlreadyAdd) {
			return nil, ErrTokenAlreadyAdd
		}
		return nil, ErrInternal
	}

	return &Output{Success: true}, nil
}

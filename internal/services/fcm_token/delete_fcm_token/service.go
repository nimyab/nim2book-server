package delete_fcm_token

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/models"
)

// FcmTokenRepository defines the interface for fcm token repository operations needed by this service
type FcmTokenRepository interface {
	DeleteFcmToken(ctx context.Context, token string, userId uuid.UUID) error
}

type Service struct {
	fcmTokenRepo FcmTokenRepository
}

var service *Service

var (
	ErrInternal = errors.New("internal error")
)

func New(fcmTokenRepo FcmTokenRepository) *Service {
	service = &Service{
		fcmTokenRepo: fcmTokenRepo,
	}
	return service
}

func (s *Service) DeleteFcmToken(input *Input, userId models.ID) (*Output, error) {
	const operation = "fcm_token.add_fcm_token.AddFcmToken"

	err := s.fcmTokenRepo.DeleteFcmToken(context.Background(), input.FcmToken, userId)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation), slog.String("fcmToken", input.FcmToken))
		return nil, ErrInternal
	}

	return &Output{Success: true}, nil
}

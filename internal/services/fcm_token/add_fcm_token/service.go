package add_fcm_token

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/internal/repositories"
)

type FcmTokenRepo interface {
	AddFcmToken(ctx context.Context, data *models.FcmToken) (*models.FcmToken, error)
}

type Service struct {
	fcmTokenRepo *repositories.FcmTokenRepository
}

var service *Service

var (
	ErrInternal        = errors.New("internal error")
	ErrTokenAlreadyAdd = errors.New("token already add")
)

func New(fcmTokenRepo *repositories.FcmTokenRepository) *Service {
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

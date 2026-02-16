package delete_fcm_token

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
)

type FcmTokenRepository interface {
	DeleteByToken(ctx context.Context, token string) error
}

type Service struct {
	fcmTokenRepo FcmTokenRepository
}

var (
	ErrInternal = errors.New("internal error")
)

func New(fcmTokenRepo FcmTokenRepository) *Service {
	return &Service{fcmTokenRepo: fcmTokenRepo}
}

func (s *Service) DeleteFcmToken(input *Input, userId domain.ID) (*Output, error) {
	const operation = "fcm_token.delete_fcm_token.DeleteFcmToken"

	err := s.fcmTokenRepo.DeleteByToken(context.Background(), input.FcmToken)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation), slog.String("fcmToken", input.FcmToken))
		return nil, ErrInternal
	}

	return &Output{Success: true}, nil
}

package add_fcm_token

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/internal/domain"
)

type FcmTokenRepository interface {
	Create(ctx context.Context, token *domain.FcmToken) (*domain.FcmToken, error)
	GetByToken(ctx context.Context, token string) (*domain.FcmToken, error)
}

type Service struct {
	fcmTokenRepo FcmTokenRepository
}

var (
	ErrInternal        = errors.New("internal error")
	ErrTokenAlreadyAdd = errors.New("token already add")
)

func New(fcmTokenRepo FcmTokenRepository) *Service {
	return &Service{fcmTokenRepo: fcmTokenRepo}
}

func (s *Service) AddFcmToken(input *Input, userId domain.ID) (*Output, error) {
	const operation = "fcm_token.add_fcm_token.AddFcmToken"

	// Проверяем, существует ли уже токен
	existingToken, err := s.fcmTokenRepo.GetByToken(context.Background(), input.FcmToken)
	if err == nil && existingToken != nil {
		slog.Warn("Token already exists", slog.String("operation", operation), slog.String("token", input.FcmToken))
		return nil, ErrTokenAlreadyAdd
	}

	fcmTokenData := &domain.FcmToken{
		Token: input.FcmToken,
		User: &domain.User{
			ID: userId,
		},
	}

	_, err = s.fcmTokenRepo.Create(context.Background(), fcmTokenData)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation), slog.Any("fcmTokenData", fcmTokenData))
		if ent.IsConstraintError(err) {
			return nil, ErrTokenAlreadyAdd
		}
		return nil, ErrInternal
	}

	return &Output{Success: true}, nil
}

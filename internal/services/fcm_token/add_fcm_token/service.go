package add_fcm_token

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/adapter/postgres"
	"github.com/nimyab/nim2book-back/internal/domain"
)

type Postgres interface {
	AddFcmToken(ctx context.Context, data *domain.FcmToken) (*domain.FcmToken, error)
}

type Service struct {
	pg Postgres
}

var service *Service

var (
	ErrInternal        = errors.New("internal error")
	ErrTokenAlreadyAdd = errors.New("token already add")
)

func New(pg Postgres) *Service {
	service = &Service{
		pg: pg,
	}
	return service
}

func (s *Service) AddFcmToken(input *Input, userId domain.Id) (*Output, error) {
	const operation = "fcm_token.add_fcm_token.AddFcmToken"

	fcmTokenData := &domain.FcmToken{
		Token:  input.FcmToken,
		UserId: userId,
	}
	_, err := s.pg.AddFcmToken(context.Background(), fcmTokenData)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation), slog.Any("fcmTokenData", fcmTokenData))
		if errors.Is(err, postgres.ErrFcmTokenAlreadyAdd) {
			return nil, ErrTokenAlreadyAdd
		}
		return nil, ErrInternal
	}

	return &Output{Success: true}, nil
}

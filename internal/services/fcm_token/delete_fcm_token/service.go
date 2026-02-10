package delete_fcm_token

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
)

type Postgres interface {
	DeleteFcmToken(ctx context.Context, token string, userId domain.Id) error
}

type Service struct {
	pg Postgres
}

var (
	ErrInternal = errors.New("internal error")
)

func New(pg Postgres) *Service {
	return &Service{pg: pg}
}

func (s *Service) DeleteFcmToken(input *Input, userId domain.Id) (*Output, error) {
	const operation = "fcm_token.add_fcm_token.AddFcmToken"

	err := s.pg.DeleteFcmToken(context.Background(), input.FcmToken, userId)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation), slog.String("fcmToken", input.FcmToken))
		return nil, ErrInternal
	}

	return &Output{Success: true}, nil
}

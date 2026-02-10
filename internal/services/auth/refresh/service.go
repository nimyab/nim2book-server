package refresh

import (
	"errors"
	"log/slog"
	"time"

	"github.com/nimyab/nim2book-back/pkg/jwt"
)

type Service struct {
	secret      string
	accessTime  time.Duration
	refreshTime time.Duration
}

func New(secret string, accessTime, refreshTime time.Duration) *Service {
	return &Service{
		secret:      secret,
		accessTime:  accessTime,
		refreshTime: refreshTime,
	}
}

var (
	ErrParseTokenFailed = errors.New("parse token failed")
	ErrInternal         = errors.New("internal error")
)

func (s *Service) Refresh(input *Input) (*Output, error) {
	const operation = "auth.refresh.Refresh"

	payload, err := jwt.ParseToken(input.RefreshToken, s.secret)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrParseTokenFailed
	}

	accessToken, refreshToken, err := jwt.GenerateTokens(
		payload,
		s.secret,
		s.accessTime,
		s.refreshTime,
	)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	return &Output{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

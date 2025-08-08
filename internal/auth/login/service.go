package login

import (
	"context"
	"errors"
	"github.com/nimyab/nim2book-back/internal/adapter/postgres"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"time"
)

type Postgres interface {
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
}

type Service struct {
	pg          Postgres
	secret      string
	accessTime  time.Duration
	refreshTime time.Duration
}

var service *Service

var (
	ErrInternal           = errors.New("internal error")
	ErrPasswordDoNotMatch = errors.New("passwords do not match")
)

func New(pg Postgres, secret string, accessTime, refreshTime time.Duration) *Service {
	service = &Service{
		pg:          pg,
		secret:      secret,
		accessTime:  accessTime,
		refreshTime: refreshTime,
	}
	return service
}

func (s *Service) Login(input *Input) (*Output, error) {
	const operation = "auth.login.Login"

	user, err := s.pg.GetUserByEmail(context.Background(), input.Email)
	if errors.Is(postgres.ErrUserNotFound, err) {
		return nil, err
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if errors.Is(bcrypt.ErrMismatchedHashAndPassword, err) {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrPasswordDoNotMatch
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	accessToken, refreshToken, err := jwt.GenerateTokens(
		domain.JwtPayload{
			Id:      user.Id,
			IsAdmin: user.IsAdmin,
		},
		s.secret,
		s.accessTime,
		s.refreshTime,
	)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	return &Output{User: user, AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

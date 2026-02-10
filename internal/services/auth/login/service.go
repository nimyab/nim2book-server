package login

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/nimyab/nim2book-back/internal/adapter/postgres_sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/helpers"
	"github.com/nimyab/nim2book-back/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Postgres interface {
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	GetPersonalUserBooks(ctx context.Context, query postgres_sqlc.GetPersonalUserBooksQuery) ([]domain.PersonalUserBook, error)
	GetPersonalUserBookGenres(ctx context.Context, bookId domain.Id) ([]domain.Genre, error)
}

type Service struct {
	pg          Postgres
	secret      string
	accessTime  time.Duration
	refreshTime time.Duration
}

var (
	ErrInternal           = errors.New("internal error")
	ErrPasswordDoNotMatch = errors.New("passwords do not match")
)

func New(pg Postgres, secret string, accessTime, refreshTime time.Duration) *Service {
	return &Service{
		pg:          pg,
		secret:      secret,
		accessTime:  accessTime,
		refreshTime: refreshTime,
	}
}

func (s *Service) Login(input *Input) (*Output, error) {
	const operation = "auth.login.Login"

	user, err := s.pg.GetUserByEmail(context.Background(), input.Email)
	if errors.Is(postgres_sqlc.ErrUserNotFound, err) {
		return nil, err
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.EmailPasswordAccount.PasswordHash), []byte(input.Password))
	if err != nil && errors.Is(bcrypt.ErrMismatchedHashAndPassword, err) {
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
			IsVIP:   user.IsVIP,
		},
		s.secret,
		s.accessTime,
		s.refreshTime,
	)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	// Загружаем персональные книги пользователя вместе с жанрами
	if err := helpers.EnrichUserWithPersonalBooksAndGenres(context.Background(), user, s.pg, operation); err != nil {
		return nil, ErrInternal
	}

	return &Output{User: user, AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

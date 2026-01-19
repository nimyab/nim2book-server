package google_login

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/nimyab/nim2book-back/internal/adapter/postgres_sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/helpers"
	"github.com/nimyab/nim2book-back/pkg/jwt"
	"google.golang.org/api/idtoken"
)

type Postgres interface {
	GetUserByGoogleSub(ctx context.Context, sub string) (*domain.User, error)
	CreateUserByGoogle(ctx context.Context, data *domain.GoogleAccount) (*domain.User, error)
	GetPersonalUserBooks(ctx context.Context, query postgres_sqlc.GetPersonalUserBooksQuery) ([]domain.PersonalUserBook, error)
	GetPersonalUserBookGenres(ctx context.Context, bookId domain.Id) ([]domain.Genre, error)
}

type Service struct {
	pg             Postgres
	secret         string
	googleClientId string
	accessTime     time.Duration
	refreshTime    time.Duration
}

var service *Service

var (
	ErrInternal          = errors.New("internal error")
	ErrInvalidToken      = errors.New("invalid token")
	ErrInvalidGoogleData = errors.New("invalid google data")
)

func New(pg Postgres, googleClientId string, secret string, accessTime, refreshTime time.Duration) *Service {
	service = &Service{
		pg:             pg,
		secret:         secret,
		accessTime:     accessTime,
		refreshTime:    refreshTime,
		googleClientId: googleClientId,
	}
	return service
}

func (s *Service) GoogleLogin(input *Input) (*Output, error) {
	const operation = "auth.login.GoogleLogin"

	payload, err := idtoken.Validate(context.Background(), input.IdToken, s.googleClientId)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInvalidToken
	}
	email, ok1 := payload.Claims["email"].(string)
	sub, ok2 := payload.Claims["sub"].(string)
	emailVerified, ok3 := payload.Claims["email_verified"].(bool)
	name, ok4 := payload.Claims["name"].(string)
	if !(ok1 && ok2 && ok3 && ok4) {
		return nil, ErrInvalidGoogleData
	}
	var picture *string = nil
	if pic, ok := payload.Claims["picture"].(string); ok {
		picture = &pic
	}

	googleUser := &domain.GoogleAccount{
		Email:         email,
		EmailVerified: emailVerified,
		Sub:           sub,
		Name:          name,
		Picture:       picture,
	}

	user, err := s.pg.GetUserByGoogleSub(context.Background(), googleUser.Sub)
	if err != nil && !errors.Is(err, postgres_sqlc.ErrUserNotFound) {
		slog.Error(err.Error(), slog.String("operation", operation), slog.Any("googleUser", googleUser))
		return nil, ErrInternal
	}
	// елси такого пользователя нет, то создаем его
	if errors.Is(err, postgres_sqlc.ErrUserNotFound) {
		user, err = s.pg.CreateUserByGoogle(context.Background(), googleUser)
		if err != nil {
			slog.Error(err.Error(), slog.String("operation", operation), slog.Any("googleUser", googleUser))
			return nil, ErrInternal
		}
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

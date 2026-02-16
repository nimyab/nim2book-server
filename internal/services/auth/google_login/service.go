package google_login

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
	"github.com/nimyab/nim2book-back/pkg/jwt"
	"google.golang.org/api/idtoken"
)

type UserRepository interface {
	GetGoogleAccountBySub(ctx context.Context, sub string) (*domain.GoogleAccount, error)
	CreateWithGoogleAccount(ctx context.Context, user *domain.User, googleAccount *domain.GoogleAccount) (*domain.User, error)
}

type Service struct {
	userRepo       UserRepository
	secret         string
	googleClientId string
	accessTime     time.Duration
	refreshTime    time.Duration
}

var (
	ErrInternal          = errors.New("internal error")
	ErrInvalidToken      = errors.New("invalid token")
	ErrInvalidGoogleData = errors.New("invalid google data")
)

func New(userRepo UserRepository, googleClientId string, secret string, accessTime, refreshTime time.Duration) *Service {
	return &Service{
		userRepo:       userRepo,
		secret:         secret,
		accessTime:     accessTime,
		refreshTime:    refreshTime,
		googleClientId: googleClientId,
	}
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
	var picture string
	if pic, ok := payload.Claims["picture"].(string); ok {
		picture = pic
	}

	googleAccount := &domain.GoogleAccount{
		Email:         email,
		EmailVerified: emailVerified,
		Sub:           sub,
		Name:          name,
		Picture:       picture,
	}

	// Проверяем, существует ли Google аккаунт
	existingAccount, err := s.userRepo.GetGoogleAccountBySub(context.Background(), googleAccount.Sub)
	var user *domain.User

	if err != nil && !errors.Is(err, repository.ErrNotFound) {
		slog.Error(err.Error(), slog.String("operation", operation), slog.Any("googleAccount", googleAccount))
		return nil, ErrInternal
	}

	// Если такого пользователя нет, то создаем его
	if existingAccount == nil || errors.Is(err, repository.ErrNotFound) {
		newUser := &domain.User{
			IsVIP:    false,
			IsAdmin:  false,
			Metadata: map[string]interface{}{},
		}
		user, err = s.userRepo.CreateWithGoogleAccount(context.Background(), newUser, googleAccount)
		if err != nil {
			slog.Error(err.Error(), slog.String("operation", operation), slog.Any("googleAccount", googleAccount))
			return nil, ErrInternal
		}
	} else {
		// Используем существующего пользователя
		user = existingAccount.User
	}

	accessToken, refreshToken, err := jwt.GenerateTokens(
		domain.JwtPayload{
			ID:      user.ID,
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

	return &Output{User: user, AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

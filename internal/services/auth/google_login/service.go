package google_login

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/internal/repositories"
	"github.com/nimyab/nim2book-back/pkg/jwt"
	"google.golang.org/api/idtoken"
)

type Service struct {
	userRepo       *repositories.UserRepository
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

func New(userRepo *repositories.UserRepository, googleClientId string, secret string, accessTime, refreshTime time.Duration) *Service {
	service = &Service{
		userRepo:       userRepo,
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

	googleUser := &models.GoogleAccount{
		Email:         email,
		EmailVerified: emailVerified,
		Sub:           sub,
		Name:          name,
		Picture:       picture,
	}

	user, err := s.userRepo.GetUserByGoogleSub(context.Background(), googleUser.Sub)
	if err != nil && !errors.Is(err, repositories.ErrUserNotFound) {
		slog.Error(err.Error(), slog.String("operation", operation), slog.Any("googleUser", googleUser))
		return nil, ErrInternal
	}
	// елси такого пользователя нет, то создаем его
	if errors.Is(err, repositories.ErrUserNotFound) {
		user, err = s.userRepo.CreateUserByGoogle(context.Background(), googleUser)
		if err != nil {
			slog.Error(err.Error(), slog.String("operation", operation), slog.Any("googleUser", googleUser))
			return nil, ErrInternal
		}
	}

	accessToken, refreshToken, err := jwt.GenerateTokens(
		models.JwtPayload{
			Id:      user.ID,
			IsAdmin: user.IsAdmin,
			IsVIP:   user.IsVip,
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

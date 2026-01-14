package login

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/internal/repositories"
	"github.com/nimyab/nim2book-back/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository defines the interface for user repository operations needed by this service
type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type Service struct {
	userRepo    UserRepository
	secret      string
	accessTime  time.Duration
	refreshTime time.Duration
}

var service *Service

var (
	ErrInternal           = errors.New("internal error")
	ErrPasswordDoNotMatch = errors.New("passwords do not match")
	ErrEmptyEmail         = errors.New("email cannot be empty")
	ErrEmptyPassword      = errors.New("password cannot be empty")
)

func New(userRepo UserRepository, secret string, accessTime, refreshTime time.Duration) *Service {
	service = &Service{
		userRepo:    userRepo,
		secret:      secret,
		accessTime:  accessTime,
		refreshTime: refreshTime,
	}
	return service
}

func (s *Service) Login(input *Input) (*Output, error) {
	const operation = "auth.login.Login"

	// Validate input
	if input.Email == "" {
		return nil, ErrEmptyEmail
	}
	if input.Password == "" {
		return nil, ErrEmptyPassword
	}

	user, err := s.userRepo.GetUserByEmail(context.Background(), input.Email)
	if errors.Is(repositories.ErrUserNotFound, err) {
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

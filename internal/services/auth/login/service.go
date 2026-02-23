package login

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
	"github.com/nimyab/nim2book-back/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	GetByBasicAccountEmail(ctx context.Context, email string) (*domain.User, error)
}

type Service struct {
	userRepo    UserRepository
	secret      string
	accessTime  time.Duration
	refreshTime time.Duration
}

var (
	ErrInternal           = errors.New("internal error")
	ErrPasswordDoNotMatch = errors.New("passwords do not match")
	ErrUserNotFound       = errors.New("user not found")
)

func New(userRepo UserRepository, secret string, accessTime, refreshTime time.Duration) *Service {
	return &Service{
		userRepo:    userRepo,
		secret:      secret,
		accessTime:  accessTime,
		refreshTime: refreshTime,
	}
}

func (s *Service) Login(ctx context.Context, input *Input) (*Output, error) {
	const operation = "auth.login.Login"

	// Получаем пользователя по email
	user, err := s.userRepo.GetByBasicAccountEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrUserNotFound
		}
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	// Проверяем наличие BasicAccount
	if user.BasicAccount == nil {
		slog.Error("User has no basic account", slog.String("operation", operation), slog.Any("userId", user.ID))
		return nil, ErrUserNotFound
	}

	// Сравниваем пароли
	err = bcrypt.CompareHashAndPassword([]byte(user.BasicAccount.PasswordHash), []byte(input.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			slog.Warn("Password mismatch", slog.String("operation", operation), slog.String("email", input.Email))
			return nil, ErrPasswordDoNotMatch
		}
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	// Генерируем токены
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

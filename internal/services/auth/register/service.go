package register

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrUserAlreadyExist = errors.New("user already exists")
	ErrInternal         = errors.New("internal error")
)

type Service struct {
	userRepo *repositories.UserRepository
}

var service *Service

func New(userRepo *repositories.UserRepository) *Service {
	service = &Service{
		userRepo: userRepo,
	}
	return service
}

func (s *Service) Register(input *Input) (*Output, error) {
	const operation = "auth.register.Register"

	user, err := s.userRepo.GetUserByEmail(context.Background(), input.Email)
	if user != nil {
		return nil, ErrUserAlreadyExist
	}
	if err != nil && !errors.Is(repositories.ErrUserNotFound, err) {
		return nil, err
	}

	passwordHashBytes, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		slog.Error(
			err.Error(),
			slog.String("password", input.Password),
			slog.String("operation", operation),
		)
		return nil, ErrInternal
	}

	user, err = s.userRepo.CreateUserByEmailAndPassword(context.Background(), input.Email, string(passwordHashBytes))
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	slog.Info("create user", slog.Any("user", user), slog.String("operation", operation))
	return &Output{Success: true}, nil
}

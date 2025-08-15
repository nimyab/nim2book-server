package register

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/adapter/postgres"
	"github.com/nimyab/nim2book-back/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type Postgres interface {
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
}

var (
	ErrUserAlreadyExist = errors.New("user already exists")
	ErrInternal         = errors.New("internal error")
)

type Service struct {
	pg Postgres
}

var service *Service

func New(pg Postgres) *Service {
	service = &Service{
		pg: pg,
	}
	return service
}

func (s *Service) Register(input *Input) (*Output, error) {
	const operation = "auth.register.Register"

	user, err := s.pg.GetUserByEmail(context.Background(), input.Email)
	if user != nil {
		return nil, ErrUserAlreadyExist
	}
	if err != nil && !errors.Is(postgres.ErrUserNotFound, err) {
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

	newUser := &domain.User{
		Email:        input.Email,
		PasswordHash: string(passwordHashBytes),
		IsAdmin:      false,
	}
	user, err = s.pg.CreateUser(context.Background(), newUser)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	slog.Info("create user", slog.Any("user", user), slog.String("operation", operation))
	return &Output{Success: true}, nil
}

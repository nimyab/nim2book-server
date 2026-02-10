package register

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/adapter/postgres_sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type Postgres interface {
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	CreateUserByEmailAndPassword(ctx context.Context, data *domain.EmailPasswordAccount) (*domain.User, error)
}

var (
	ErrUserAlreadyExist = errors.New("user already exists")
	ErrInternal         = errors.New("internal error")
)

type Service struct {
	pg Postgres
}

func New(pg Postgres) *Service {
	return &Service{pg: pg}
}

func (s *Service) Register(input *Input) (*Output, error) {
	const operation = "auth.register.Register"

	user, err := s.pg.GetUserByEmail(context.Background(), input.Email)
	if user != nil {
		return nil, ErrUserAlreadyExist
	}
	if err != nil && !errors.Is(postgres_sqlc.ErrUserNotFound, err) {
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

	newUser := &domain.EmailPasswordAccount{
		Email:        input.Email,
		PasswordHash: string(passwordHashBytes),
	}
	user, err = s.pg.CreateUserByEmailAndPassword(context.Background(), newUser)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	slog.Info("create user", slog.Any("user", user), slog.String("operation", operation))
	return &Output{Success: true}, nil
}

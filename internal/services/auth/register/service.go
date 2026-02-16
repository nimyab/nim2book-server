package register

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	GetBasicAccountByEmail(ctx context.Context, email string) (*domain.BasicAccount, error)
	CreateWithBasicAccount(ctx context.Context, user *domain.User, basicAccount *domain.BasicAccount) (*domain.User, error)
}

var (
	ErrUserAlreadyExist = errors.New("user already exists")
	ErrInternal         = errors.New("internal error")
)

type Service struct {
	userRepo UserRepository
}

func New(userRepo UserRepository) *Service {
	return &Service{userRepo: userRepo}
}

func (s *Service) Register(ctx context.Context, input *Input) (*Output, error) {
	const operation = "auth.register.Register"

	// Проверяем, существует ли пользователь с таким email
	existingAccount, err := s.userRepo.GetBasicAccountByEmail(ctx, input.Email)
	if err == nil && existingAccount != nil {
		return nil, ErrUserAlreadyExist
	}
	// Если ошибка не NotFound - возвращаем ошибку
	if err != nil && !ent.IsNotFound(err) && !errors.Is(err, repository.ErrNotFound) {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
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

	newBasicAccount := &domain.BasicAccount{
		Email:        input.Email,
		PasswordHash: string(passwordHashBytes),
		IsVerified:   false,
		VerifyLink:   "",
	}

	newUser := &domain.User{
		IsVIP:    false,
		IsAdmin:  false,
		Metadata: map[string]interface{}{},
	}

	user, err := s.userRepo.CreateWithBasicAccount(ctx, newUser, newBasicAccount)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternal
	}

	slog.Info("create user", slog.Any("user", user), slog.String("operation", operation))
	return &Output{Success: true}, nil
}

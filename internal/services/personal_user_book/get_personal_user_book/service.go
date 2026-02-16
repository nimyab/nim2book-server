package get_personal_user_book

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
)

var (
	ErrBookNotFound = errors.New("book not found")
	ErrForbidden    = errors.New("you don't have access to this book")
)

type PersonalBookRepository interface {
	GetByID(ctx context.Context, id domain.ID) (*domain.PersonalBook, error)
}

type Service struct {
	personalBookRepo PersonalBookRepository
}

func New(personalBookRepo PersonalBookRepository) *Service {
	return &Service{personalBookRepo: personalBookRepo}
}

func (s *Service) GetPersonalUserBook(input *Input) (*Output, error) {
	const operation = "personal_user_book.get_personal_user_book.GetPersonalUserBook"

	book, err := s.personalBookRepo.GetByID(context.Background(), input.BookId)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrBookNotFound
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	// Проверяем, что книга принадлежит пользователю
	if book.User == nil || book.User.ID != input.UserId {
		slog.Warn("user tried to access book that doesn't belong to them",
			slog.String("operation", operation),
			slog.String("userId", input.UserId.String()),
			slog.String("bookId", input.BookId.String()),
		)
		return nil, ErrForbidden
	}

	return &Output{Book: book}, nil
}

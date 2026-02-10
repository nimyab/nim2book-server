package get_personal_user_book

import (
	"context"
	"errors"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/adapter/postgres_sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/helpers"
)

var (
	ErrBookNotFound = errors.New("book not found")
	ErrForbidden    = errors.New("you don't have access to this book")
)

type Postgres interface {
	GetPersonalUserBook(ctx context.Context, id domain.Id) (*domain.PersonalUserBook, error)
	GetPersonalUserBookGenres(ctx context.Context, bookId domain.Id) ([]domain.Genre, error)
}

type Service struct {
	pg Postgres
}

func New(pg Postgres) *Service {
	return &Service{pg: pg}
}

func (s *Service) GetPersonalUserBook(input *Input) (*Output, error) {
	const operation = "personal_user_book.get_personal_user_book.GetPersonalUserBook"

	book, err := s.pg.GetPersonalUserBook(context.Background(), input.BookId)
	if errors.Is(err, postgres_sqlc.ErrPersonalUserBookNotFound) {
		return nil, ErrBookNotFound
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	// Проверяем, что книга принадлежит пользователю
	if book.UserId != input.UserId {
		slog.Warn("user tried to access book that doesn't belong to them",
			slog.String("operation", operation),
			slog.String("userId", input.UserId.String()),
			slog.String("bookId", input.BookId.String()),
			slog.String("bookOwnerId", book.UserId.String()),
		)
		return nil, ErrForbidden
	}

	// Загружаем жанры книги
	if err := helpers.EnrichPersonalBookWithGenres(context.Background(), book, s.pg, operation); err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	return &Output{Book: book}, nil
}

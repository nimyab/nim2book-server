package get_personal_user_books

import (
	"context"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/adapter/postgres_sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/helpers"
)

type Postgres interface {
	GetPersonalUserBooks(ctx context.Context, query postgres_sqlc.GetPersonalUserBooksQuery) ([]domain.PersonalUserBook, error)
	GetPersonalUserBookGenres(ctx context.Context, bookId domain.Id) ([]domain.Genre, error)
}

type Service struct {
	pg Postgres
}

func New(pg Postgres) *Service {
	return &Service{pg: pg}
}

func (s *Service) GetPersonalUserBooks(input *Input) (*Output, error) {
	const operation = "personal_user_book.get_personal_user_books.GetPersonalUserBooks"

	books, err := s.pg.GetPersonalUserBooks(context.Background(), postgres_sqlc.GetPersonalUserBooksQuery{
		UserId:  input.UserId,
		Author:  input.Author,
		Title:   input.Title,
		GenreId: input.GenreId,
		Page:    input.Page,
	})
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	// Загружаем жанры для каждой книги
	if err := helpers.EnrichPersonalBooksWithGenres(context.Background(), books, s.pg, operation); err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	return &Output{Books: books}, nil
}

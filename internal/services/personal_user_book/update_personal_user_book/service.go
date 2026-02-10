package update_personal_user_book

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/adapter/postgres_sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/helpers"
)

var (
	ErrFailedToOpenFile   = errors.New("failed to open file")
	ErrFailedToReadFile   = errors.New("failed to read file")
	ErrFailedToUploadFile = errors.New("failed to upload book")
	ErrBookNotFound       = errors.New("book not found")
	ErrForbidden          = errors.New("you don't have access to this book")
	ErrInternalServer     = errors.New("internal server error")
)

type S3 interface {
	Upload(path string, data []byte) error
}

type Postgres interface {
	GetPersonalUserBook(ctx context.Context, id domain.Id) (*domain.PersonalUserBook, error)
	UpdatePersonalUserBook(ctx context.Context, book *domain.PersonalUserBook) error
	GetPersonalUserBookGenres(ctx context.Context, bookId domain.Id) ([]domain.Genre, error)
}

type Service struct {
	pg Postgres
	s3 S3
}

func New(pg Postgres, s3 S3) *Service {
	return &Service{pg: pg, s3: s3}
}

func (s *Service) UpdatePersonalUserBook(input *Input, cover *multipart.FileHeader) (*Output, error) {
	const operation = "personal_user_book.update_personal_user_book.UpdatePersonalUserBook"

	book, err := s.pg.GetPersonalUserBook(context.Background(), input.Id)
	if errors.Is(err, postgres_sqlc.ErrPersonalUserBookNotFound) {
		return nil, ErrBookNotFound
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrBookNotFound
	}

	// Проверяем, что книга принадлежит пользователю
	if book.UserId != input.UserId {
		slog.Warn("user tried to update book that doesn't belong to them",
			slog.String("operation", operation),
			slog.String("userId", input.UserId.String()),
			slog.String("bookId", input.Id.String()),
			slog.String("bookOwnerId", book.UserId.String()),
		)
		return nil, ErrForbidden
	}

	if cover != nil {
		file, err := cover.Open()
		if err != nil {
			slog.Error(err.Error(), slog.String("operation", operation))
			return nil, ErrFailedToOpenFile
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			slog.Error(err.Error(), slog.String("operation", operation))
			return nil, ErrFailedToReadFile
		}

		path := fmt.Sprintf(
			"cover/%s/%s%s",
			strings.ReplaceAll(book.Title, " ", "_"),
			uuid.New().String(),
			filepath.Ext(cover.Filename),
		)
		if err = s.s3.Upload(path, data); err != nil {
			slog.Error(err.Error(), slog.String("operation", operation))
			return nil, ErrFailedToUploadFile
		}

		book.Cover = &path
	}

	if input.Title != nil {
		book.Title = *input.Title
	}

	if input.Author != nil {
		book.Author = *input.Author
	}

	if err = s.pg.UpdatePersonalUserBook(context.Background(), book); err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternalServer
	}

	// Загружаем жанры книги
	if err := helpers.EnrichPersonalBookWithGenres(context.Background(), book, s.pg, operation); err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternalServer
	}

	return &Output{Book: book}, nil
}

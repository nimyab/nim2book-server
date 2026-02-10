package update_book

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
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/helpers"
)

var (
	ErrFailedToOpenFile   = errors.New("failed to open file")
	ErrFailedToReadFile   = errors.New("failed to read file")
	ErrFailedToUploadFile = errors.New("failed to upload book")
	ErrBookNotFound       = errors.New("book not found")
	ErrInternalServer     = errors.New("internal server error")
)

type S3 interface {
	Upload(path string, data []byte) error
}

type Postgres interface {
	GetBook(ctx context.Context, id domain.Id) (*domain.Book, error)
	UpdateBook(ctx context.Context, book *domain.Book) error
	GetBookGenres(ctx context.Context, bookId domain.Id) ([]domain.Genre, error)
}

type Service struct {
	pg Postgres
	s3 S3
}

func New(pg Postgres, s3 S3) *Service {
	return &Service{
		pg: pg,
		s3: s3,
	}
}

func (s *Service) UpdateBook(input *Input, cover *multipart.FileHeader) (*Output, error) {
	const operation = "book.update_book.UpdateBook"

	book, err := s.pg.GetBook(context.Background(), input.Id)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrBookNotFound
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

	if err = s.pg.UpdateBook(context.Background(), book); err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternalServer
	}

	// Загружаем жанры книги
	if err := helpers.EnrichBookWithGenres(context.Background(), book, s.pg, operation); err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternalServer
	}

	return &Output{Book: book}, nil
}

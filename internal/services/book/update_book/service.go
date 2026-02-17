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

type BookRepository interface {
	GetByID(ctx context.Context, id domain.ID) (*domain.Book, error)
	Update(ctx context.Context, book *domain.Book) (*domain.Book, error)
}

type AuthorRepository interface {
	GetOrCreate(ctx context.Context, name string) (*domain.Author, error)
}

type Service struct {
	bookRepo   BookRepository
	authorRepo AuthorRepository
	s3         S3
}

func New(bookRepo BookRepository, authorRepo AuthorRepository, s3 S3) *Service {
	return &Service{
		bookRepo:   bookRepo,
		authorRepo: authorRepo,
		s3:         s3,
	}
}

func (s *Service) UpdateBook(ctx context.Context, input *Input, cover *multipart.FileHeader) (*Output, error) {
	const operation = "book.update_book.UpdateBook"

	book, err := s.bookRepo.GetByID(ctx, input.Id)
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

		book.CoverURL = &path
	}

	if input.Title != nil {
		book.Title = *input.Title
	}

	if input.Author != nil {
		author, err := s.authorRepo.GetOrCreate(ctx, *input.Author)
		if err != nil {
			slog.Error(err.Error(), slog.String("operation", operation))
			return nil, ErrInternalServer
		}
		book.Author = author
	}

	updatedBook, err := s.bookRepo.Update(ctx, book)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternalServer
	}

	return &Output{Book: updatedBook}, nil
}

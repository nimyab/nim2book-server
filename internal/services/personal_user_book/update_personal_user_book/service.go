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
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/internal/repository"
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

type PersonalBookRepository interface {
	GetByID(ctx context.Context, id domain.ID) (*domain.PersonalBook, error)
	Update(ctx context.Context, book *domain.PersonalBook) (*domain.PersonalBook, error)
}

type AuthorRepository interface {
	GetOrCreate(ctx context.Context, name string) (*domain.Author, error)
}

type Service struct {
	personalBookRepo PersonalBookRepository
	authorRepo       AuthorRepository
	s3               S3
}

func New(personalBookRepo PersonalBookRepository, authorRepo AuthorRepository, s3 S3) *Service {
	return &Service{
		personalBookRepo: personalBookRepo,
		authorRepo:       authorRepo,
		s3:               s3,
	}
}

func (s *Service) UpdatePersonalUserBook(ctx context.Context, input *Input, cover *multipart.FileHeader) (*Output, error) {
	const operation = "personal_user_book.update_personal_user_book.UpdatePersonalUserBook"

	book, err := s.personalBookRepo.GetByID(ctx, input.Id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrBookNotFound
	}
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrBookNotFound
	}

	// Проверяем, что книга принадлежит пользователю
	if book.User == nil || book.User.ID != input.UserId {
		slog.Warn("user tried to update book that doesn't belong to them",
			slog.String("operation", operation),
			slog.String("userId", input.UserId.String()),
			slog.String("bookId", input.Id.String()),
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

	updatedBook, err := s.personalBookRepo.Update(ctx, book)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, ErrInternalServer
	}

	return &Output{Book: updatedBook}, nil
}

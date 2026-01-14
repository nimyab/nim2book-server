package get_book

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/models"
)

// BookRepository defines the interface for book repository operations needed by this service
type BookRepository interface {
	GetBookById(ctx context.Context, id uuid.UUID) (*models.Book, error)
}

type Service struct {
	bookRepo BookRepository
}

var service *Service

func New(bookRepo BookRepository) *Service {
	service = &Service{
		bookRepo: bookRepo,
	}
	return service
}

func (s *Service) GetBook(input *Input) (*Output, error) {
	const operation = "book.get_book.GetBook"

	book, err := s.bookRepo.GetBookById(context.Background(), input.Id)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	output := &Output{Book: book}
	return output, nil
}

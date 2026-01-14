package get_books

import (
	"context"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/models"
)

// BookRepository defines the interface for book repository operations needed by this service
type BookRepository interface {
	GetBooks(ctx context.Context, author, title string, page int) ([]*models.Book, error)
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

func (s *Service) GetBooks(input *Input) (*Output, error) {
	const operation = "book.get_books.GetBooks"

	books, err := s.bookRepo.GetBooks(context.Background(), input.Author, input.Title, input.Page)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	result := make([]models.Book, len(books))
	for i, book := range books {
		result[i] = *book
	}

	return &Output{Books: result}, nil
}

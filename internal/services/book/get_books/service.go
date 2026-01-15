package get_books

import (
	"context"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/internal/repositories"
)

type Service struct {
	bookRepo *repositories.BookRepository
}

var service *Service

func New(bookRepo *repositories.BookRepository) *Service {
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

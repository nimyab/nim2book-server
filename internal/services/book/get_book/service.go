package get_book

import (
	"context"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/domain"
)

type BookRepository interface {
	GetByID(ctx context.Context, id domain.ID) (*domain.Book, error)
}

type Service struct {
	bookRepo BookRepository
}

func New(bookRepo BookRepository) *Service {
	return &Service{
		bookRepo: bookRepo,
	}
}

func (s *Service) GetBook(ctx context.Context, input *Input) (*Output, error) {
	const operation = "book.get_book.GetBook"

	book, err := s.bookRepo.GetByID(ctx, input.Id)
	if err != nil {
		slog.Error(err.Error(), slog.String("operation", operation))
		return nil, err
	}

	output := &Output{Book: book}
	return output, nil
}

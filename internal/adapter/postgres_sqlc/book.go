package postgres_sqlc

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/nimyab/nim2book-back/db/sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/transaction"
)

var (
	ErrBookNotFound      = errors.New("book not found")
	ErrBookAlreadyExists = errors.New("book already exists")
)

const (
	step = 100
)

type GetBooksQuery struct {
	Author string
	Title  string
	Page   int
}

func (db *Postgres) GetBookByAuthorAndTitle(ctx context.Context, author, title string) (*domain.Book, error) {
	const operation = "postgres_sqlc.GetBookByAuthorAndTitle"

	book, err := db.Queries.GetBookByAuthorAndTitle(ctx, sqlc.GetBookByAuthorAndTitleParams{
		Author: author,
		Title:  title,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrBookNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return bookFromSqlc(book), nil
}

func (db *Postgres) CreateBook(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	const operation = "postgres_sqlc.CreateBook"

	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (*domain.Book, error) {
		queries := db.Queries.WithTx(tx)

		// Проверяем существует ли уже книга с таким автором и названием
		existedBook, err := queries.GetBookByAuthorAndTitle(ctx, sqlc.GetBookByAuthorAndTitleParams{
			Author: book.Author,
			Title:  book.Title,
		})
		if err == nil {
			return bookFromSqlc(existedBook), ErrBookAlreadyExists
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		// Создаем книгу
		id, err := queries.CreateBook(ctx, sqlc.CreateBookParams{
			Title:        book.Title,
			Author:       book.Author,
			ChapterPaths: book.ChapterPaths,
			Cover:        book.Cover,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		book.Id = id
		return book, nil
	})
}

func (db *Postgres) GetBook(ctx context.Context, id domain.Id) (*domain.Book, error) {
	const operation = "postgres_sqlc.GetBook"

	book, err := db.Queries.GetBookById(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrBookNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return bookFromSqlc(book), nil
}

func (db *Postgres) GetBooks(ctx context.Context, query GetBooksQuery) ([]domain.Book, error) {
	const operation = "postgres_sqlc.GetBooks"

	books, err := db.Queries.GetBooks(ctx, sqlc.GetBooksParams{
		Author: &query.Author,
		Title:  &query.Title,
		Limit:  int32(step),
		Offset: int32((query.Page - 1) * step),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	result := make([]domain.Book, len(books))
	for i, book := range books {
		result[i] = *bookFromSqlc(book)
	}

	return result, nil
}

func (db *Postgres) UpdateBook(ctx context.Context, book *domain.Book) error {
	const operation = "postgres_sqlc.UpdateBook"

	return transaction.Tx(ctx, db.Pool, func(tx pgx.Tx) error {
		queries := db.Queries.WithTx(tx)

		err := queries.UpdateBook(ctx, sqlc.UpdateBookParams{
			Title:  book.Title,
			Author: book.Author,
			Cover:  book.Cover,
			ID:     book.Id,
		})
		if err != nil {
			return fmt.Errorf("%s: %w", operation, err)
		}

		return nil
	})
}

// Конвертирует sqlc.Book в domain.Book
func bookFromSqlc(book sqlc.Book) *domain.Book {
	return &domain.Book{
		Id:           book.ID,
		Title:        book.Title,
		Author:       book.Author,
		ChapterPaths: book.ChapterPaths,
		Cover:        book.Cover,
	}
}

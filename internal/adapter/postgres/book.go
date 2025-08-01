package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nimyab/nim2book-back/internal/domain"
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
	const operation = "postgres.GetBookByAuthorAndTitle"

	sql := `SELECT * FROM books WHERE author = $1 AND title = $2`

	book := new(domain.Book)
	err := db.Pool.QueryRow(ctx, sql, author, title).Scan(book)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrBookNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return book, nil
}

func (db *Postgres) CreateBook(ctx context.Context, book *domain.Book) (*domain.Book, error) {
	const operation = "postgres.CreateBook"

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: unable to begin transaction: %w", operation, err)
	}
	defer tx.Rollback(ctx)

	sql := `select * from books where author = $1 and title = $2`
	existedBook := new(domain.Book)
	err = tx.QueryRow(ctx, sql, book.Author, book.Title).Scan(existedBook)
	if err == nil {
		return existedBook, ErrBookAlreadyExists
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	sql = `insert into books (title, author, chapter_paths) values ($1, $2, $3) returning id`
	var id uuid.UUID
	err = tx.QueryRow(ctx, sql, book.Title, book.Author, book.ChapterPaths).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	book.Id = id
	return book, nil
}

func (db *Postgres) GetBook(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	const operation = "postgres.GetBook"

	sql := `select * from books where id = $1`
	book := new(domain.Book)
	err := db.Pool.QueryRow(ctx, sql, id).Scan(book)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrBookNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return book, nil
}

func (db *Postgres) GetBooks(ctx context.Context, query GetBooksQuery) ([]domain.Book, error) {
	const operation = "postgres.GetBooks"

	sql := `select * from books
			where (author ilike '%' || @author || '%') and
				  (title ilike '%' || @title || '%')
			limit @limit
			offset @offset`

	args := pgx.NamedArgs{
		"author": query.Author,
		"title":  query.Title,
		"limit":  step,
		"offset": (query.Page - 1) * step,
	}

	rows, err := db.Pool.Query(ctx, sql, args)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer rows.Close()

	books, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Book])
	if errors.Is(err, pgx.ErrNoRows) {
		return []domain.Book{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return books, nil
}

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
	ErrPersonalUserBookNotFound      = errors.New("personal user book not found")
	ErrPersonalUserBookAlreadyExists = errors.New("personal user book already exists")
)

type GetPersonalUserBooksQuery struct {
	UserId domain.Id
	Author string
	Title  string
	Page   int
}

func (db *Postgres) GetPersonalUserBookByAuthorAndTitle(ctx context.Context, userId domain.Id, author, title string) (*domain.PersonalUserBook, error) {
	const operation = "postgres_sqlc.GetPersonalUserBookByAuthorAndTitle"

	book, err := db.Queries.GetPersonalUserBookByAuthorAndTitle(ctx, sqlc.GetPersonalUserBookByAuthorAndTitleParams{
		Author: author,
		Title:  title,
		UserID: uuidToPgtype(userId),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPersonalUserBookNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return personalUserBookFromSqlc(book), nil
}

func (db *Postgres) CreatePersonalUserBook(ctx context.Context, book *domain.PersonalUserBook) (*domain.PersonalUserBook, error) {
	const operation = "postgres_sqlc.CreatePersonalUserBook"

	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (*domain.PersonalUserBook, error) {
		queries := db.Queries.WithTx(tx)

		// Проверяем существует ли уже книга с таким автором и названием у данного пользователя
		existedBook, err := queries.GetPersonalUserBookByAuthorAndTitle(ctx, sqlc.GetPersonalUserBookByAuthorAndTitleParams{
			Author: book.Author,
			Title:  book.Title,
			UserID: uuidToPgtype(book.UserId),
		})
		if err == nil {
			return personalUserBookFromSqlc(existedBook), ErrPersonalUserBookAlreadyExists
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		// Создаем книгу
		createdBook, err := queries.CreatePersonalUserBook(ctx, sqlc.CreatePersonalUserBookParams{
			Title:        book.Title,
			Author:       book.Author,
			ChapterPaths: book.ChapterPaths,
			Cover:        book.Cover,
			UserID:       uuidToPgtype(book.UserId),
		})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		return personalUserBookFromSqlc(createdBook), nil
	})
}

func (db *Postgres) GetPersonalUserBook(ctx context.Context, id domain.Id) (*domain.PersonalUserBook, error) {
	const operation = "postgres_sqlc.GetPersonalUserBook"

	book, err := db.Queries.GetPersonalUserBookById(ctx, uuidToPgtype(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrPersonalUserBookNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return personalUserBookFromSqlc(book), nil
}

func (db *Postgres) GetPersonalUserBooks(ctx context.Context, query GetPersonalUserBooksQuery) ([]domain.PersonalUserBook, error) {
	const operation = "postgres_sqlc.GetPersonalUserBooks"

	books, err := db.Queries.GetPersonalUserBooksByUserId(ctx, sqlc.GetPersonalUserBooksByUserIdParams{
		UserID: uuidToPgtype(query.UserId),
		Author: &query.Author,
		Title:  &query.Title,
		Limit:  int32(step),
		Offset: int32((query.Page - 1) * step),
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	result := make([]domain.PersonalUserBook, len(books))
	for i, book := range books {
		result[i] = *personalUserBookFromSqlc(book)
	}

	return result, nil
}

func (db *Postgres) UpdatePersonalUserBook(ctx context.Context, book *domain.PersonalUserBook) error {
	const operation = "postgres_sqlc.UpdatePersonalUserBook"

	return transaction.Tx(ctx, db.Pool, func(tx pgx.Tx) error {
		queries := db.Queries.WithTx(tx)

		err := queries.UpdatePersonalUserBook(ctx, sqlc.UpdatePersonalUserBookParams{
			Title:  book.Title,
			Author: book.Author,
			Cover:  book.Cover,
			ID:     uuidToPgtype(book.Id),
			UserID: uuidToPgtype(book.UserId),
		})
		if err != nil {
			return fmt.Errorf("%s: %w", operation, err)
		}

		return nil
	})
}

func (db *Postgres) DeletePersonalUserBook(ctx context.Context, id domain.Id, userId domain.Id) error {
	const operation = "postgres_sqlc.DeletePersonalUserBook"

	return transaction.Tx(ctx, db.Pool, func(tx pgx.Tx) error {
		queries := db.Queries.WithTx(tx)

		err := queries.DeletePersonalUserBook(ctx, sqlc.DeletePersonalUserBookParams{
			ID:     uuidToPgtype(id),
			UserID: uuidToPgtype(userId),
		})
		if err != nil {
			return fmt.Errorf("%s: %w", operation, err)
		}

		return nil
	})
}

func (db *Postgres) AddGenreToPersonalUserBook(ctx context.Context, bookId domain.Id, genreId domain.Id) error {
	const operation = "postgres_sqlc.AddGenreToPersonalUserBook"

	return transaction.Tx(ctx, db.Pool, func(tx pgx.Tx) error {
		queries := db.Queries.WithTx(tx)

		err := queries.AddGenreToPersonalUserBook(ctx, sqlc.AddGenreToPersonalUserBookParams{
			PersonalUserBookID: uuidToPgtype(bookId),
			GenreID:            uuidToPgtype(genreId),
		})
		if err != nil {
			return fmt.Errorf("%s: %w", operation, err)
		}

		return nil
	})
}

func (db *Postgres) RemoveGenreFromPersonalUserBook(ctx context.Context, bookId domain.Id, genreId domain.Id) error {
	const operation = "postgres_sqlc.RemoveGenreFromPersonalUserBook"

	return transaction.Tx(ctx, db.Pool, func(tx pgx.Tx) error {
		queries := db.Queries.WithTx(tx)

		err := queries.RemoveGenreFromPersonalUserBook(ctx, sqlc.RemoveGenreFromPersonalUserBookParams{
			PersonalUserBookID: uuidToPgtype(bookId),
			GenreID:            uuidToPgtype(genreId),
		})
		if err != nil {
			return fmt.Errorf("%s: %w", operation, err)
		}

		return nil
	})
}

func (db *Postgres) GetPersonalUserBookGenres(ctx context.Context, bookId domain.Id) ([]domain.Genre, error) {
	const operation = "postgres_sqlc.GetPersonalUserBookGenres"

	genres, err := db.Queries.GetPersonalUserBookGenres(ctx, uuidToPgtype(bookId))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	result := make([]domain.Genre, len(genres))
	for i, genre := range genres {
		result[i] = domain.Genre{
			Id:   uuidFromPgtype(genre.ID),
			Name: genre.Name,
		}
	}

	return result, nil
}

// Конвертирует sqlc.PersonalUserBook в domain.PersonalUserBook
func personalUserBookFromSqlc(book sqlc.PersonalUserBook) *domain.PersonalUserBook {
	return &domain.PersonalUserBook{
		Id:           uuidFromPgtype(book.ID),
		Title:        book.Title,
		Author:       book.Author,
		ChapterPaths: book.ChapterPaths,
		Cover:        book.Cover,
		UserId:       uuidFromPgtype(book.UserID),
	}
}

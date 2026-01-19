package helpers

import (
	"context"
	"log/slog"

	"github.com/nimyab/nim2book-back/internal/adapter/postgres_sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
)

// PersonalBookLoader интерфейс для загрузки персональных книг
type PersonalBookLoader interface {
	GetPersonalUserBooks(ctx context.Context, query postgres_sqlc.GetPersonalUserBooksQuery) ([]domain.PersonalUserBook, error)
}

// BookGenreLoader интерфейс для загрузки жанров книги
type BookGenreLoader interface {
	GetBookGenres(ctx context.Context, bookId domain.Id) ([]domain.Genre, error)
}

// PersonalBookGenreLoader интерфейс для загрузки жанров персональной книги
type PersonalBookGenreLoader interface {
	GetPersonalUserBookGenres(ctx context.Context, bookId domain.Id) ([]domain.Genre, error)
}

// PersonalBookLoaderWithGenres интерфейс для загрузки персональных книг с жанрами
type PersonalBookLoaderWithGenres interface {
	PersonalBookLoader
	PersonalBookGenreLoader
}

// EnrichUserWithPersonalBooksAndGenres загружает персональные книги пользователя вместе с жанрами
// Это объединенный метод, который сразу загружает и книги, и жанры за один вызов
func EnrichUserWithPersonalBooksAndGenres(ctx context.Context, user *domain.User, loader PersonalBookLoaderWithGenres, operation string) error {
	if user == nil {
		return nil
	}

	// Загружаем персональные книги
	books, err := loader.GetPersonalUserBooks(ctx, postgres_sqlc.GetPersonalUserBooksQuery{
		UserId: user.Id,
		Author: "",
		Title:  "",
		Page:   1,
	})
	if err != nil {
		slog.Error("failed to load personal books",
			slog.String("operation", operation),
			slog.String("userId", user.Id.String()),
			slog.String("error", err.Error()),
		)
		return err
	}

	user.PersonalUserBooks = books

	// Загружаем жанры для всех книг
	if err := EnrichPersonalBooksWithGenres(ctx, user.PersonalUserBooks, loader, operation); err != nil {
		return err
	}

	return nil
}

// EnrichBookWithGenres загружает жанры книги и добавляет их в объект Book
func EnrichBookWithGenres(ctx context.Context, book *domain.Book, loader BookGenreLoader, operation string) error {
	if book == nil {
		return nil
	}

	genres, err := loader.GetBookGenres(ctx, book.Id)
	if err != nil {
		slog.Error("failed to load book genres",
			slog.String("operation", operation),
			slog.String("bookId", book.Id.String()),
			slog.String("error", err.Error()),
		)
		return err
	}

	book.Genres = genres
	return nil
}

// EnrichBooksWithGenres загружает жанры для списка книг
func EnrichBooksWithGenres(ctx context.Context, books []domain.Book, loader BookGenreLoader, operation string) error {
	for i := range books {
		if err := EnrichBookWithGenres(ctx, &books[i], loader, operation); err != nil {
			return err
		}
	}
	return nil
}

// EnrichPersonalBookWithGenres загружает жанры персональной книги и добавляет их в объект PersonalUserBook
func EnrichPersonalBookWithGenres(ctx context.Context, book *domain.PersonalUserBook, loader PersonalBookGenreLoader, operation string) error {
	if book == nil {
		return nil
	}

	genres, err := loader.GetPersonalUserBookGenres(ctx, book.Id)
	if err != nil {
		slog.Error("failed to load personal book genres",
			slog.String("operation", operation),
			slog.String("bookId", book.Id.String()),
			slog.String("error", err.Error()),
		)
		return err
	}

	book.Genres = genres
	return nil
}

// EnrichPersonalBooksWithGenres загружает жанры для списка персональных книг
func EnrichPersonalBooksWithGenres(ctx context.Context, books []domain.PersonalUserBook, loader PersonalBookGenreLoader, operation string) error {
	for i := range books {
		if err := EnrichPersonalBookWithGenres(ctx, &books[i], loader, operation); err != nil {
			return err
		}
	}
	return nil
}

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
	ErrGenreNotFound      = errors.New("genre not found")
	ErrGenreAlreadyExists = errors.New("genre already exists")
)

func (db *Postgres) GetGenreById(ctx context.Context, id domain.Id) (*domain.Genre, error) {
	const operation = "postgres_sqlc.GetGenreById"

	genre, err := db.Queries.GetGenreById(ctx, uuidToPgtype(id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrGenreNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return genreFromSqlc(genre), nil
}

func (db *Postgres) GetGenreByName(ctx context.Context, name string) (*domain.Genre, error) {
	const operation = "postgres_sqlc.GetGenreByName"

	genre, err := db.Queries.GetGenreByName(ctx, name)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrGenreNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return genreFromSqlc(genre), nil
}

func (db *Postgres) GetAllGenres(ctx context.Context) ([]domain.Genre, error) {
	const operation = "postgres_sqlc.GetAllGenres"

	genres, err := db.Queries.GetAllGenres(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	result := make([]domain.Genre, len(genres))
	for i, genre := range genres {
		result[i] = *genreFromSqlc(genre)
	}

	return result, nil
}

func (db *Postgres) CreateGenre(ctx context.Context, genre *domain.Genre) (*domain.Genre, error) {
	const operation = "postgres_sqlc.CreateGenre"

	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (*domain.Genre, error) {
		queries := db.Queries.WithTx(tx)

		// Проверяем существует ли уже жанр с таким названием
		existedGenre, err := queries.GetGenreByName(ctx, genre.Name)
		if err == nil {
			return genreFromSqlc(existedGenre), ErrGenreAlreadyExists
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		// Создаем жанр
		createdGenre, err := queries.CreateGenre(ctx, genre.Name)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		return genreFromSqlc(createdGenre), nil
	})
}

func (db *Postgres) UpdateGenre(ctx context.Context, genre *domain.Genre) (*domain.Genre, error) {
	const operation = "postgres_sqlc.UpdateGenre"

	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (*domain.Genre, error) {
		queries := db.Queries.WithTx(tx)

		// Проверяем существует ли жанр
		_, err := queries.GetGenreById(ctx, uuidToPgtype(genre.Id))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrGenreNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		// Проверяем не занято ли новое название другим жанром
		existedGenre, err := queries.GetGenreByName(ctx, genre.Name)
		if err == nil && uuidFromPgtype(existedGenre.ID) != genre.Id {
			return nil, ErrGenreAlreadyExists
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		// Обновляем жанр
		updatedGenre, err := queries.UpdateGenre(ctx, sqlc.UpdateGenreParams{
			ID:   uuidToPgtype(genre.Id),
			Name: genre.Name,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		return genreFromSqlc(updatedGenre), nil
	})
}

func (db *Postgres) DeleteGenre(ctx context.Context, id domain.Id) error {
	const operation = "postgres_sqlc.DeleteGenre"

	return transaction.Tx(ctx, db.Pool, func(tx pgx.Tx) error {
		queries := db.Queries.WithTx(tx)

		err := queries.DeleteGenre(ctx, uuidToPgtype(id))
		if err != nil {
			return fmt.Errorf("%s: %w", operation, err)
		}

		return nil
	})
}

// Конвертирует sqlc.Genre в domain.Genre
func genreFromSqlc(genre sqlc.Genre) *domain.Genre {
	return &domain.Genre{
		Id:   uuidFromPgtype(genre.ID),
		Name: genre.Name,
	}
}

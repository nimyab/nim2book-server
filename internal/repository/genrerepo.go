package repository

import (
	"context"

	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/ent/genre"
	"github.com/nimyab/nim2book-back/internal/domain"
)

// GenreRepository реализует domain.GenreRepository
type GenreRepository struct {
	client *ent.Client
}

// NewGenreRepository создает новый репозиторий жанров
func NewGenreRepository(client *ent.Client) *GenreRepository {
	return &GenreRepository{client: client}
}

// getByIDInternal возвращает жанр по ID, работает как с транзакцией, так и без неё
func (r *GenreRepository) getByIDInternal(ctx context.Context, tx *ent.Tx, id domain.ID) (*domain.Genre, error) {
	client := GetClientOrTx(r.client, tx)

	entGenre, err := client.Genre.Query().
		Where(genre.ID(id)).
		Only(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return MapGenreToDomain(entGenre), nil
}

// Create создает новый жанр
func (r *GenreRepository) Create(ctx context.Context, domainGenre *domain.Genre) (*domain.Genre, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.Genre, error) {
		entGenre, err := tx.Genre.Create().
			SetName(domainGenre.Name).
			Save(ctx)

		if err != nil {
			return nil, HandleError(err)
		}

		return r.getByIDInternal(ctx, tx, entGenre.ID)
	})
}

// GetByID возвращает жанр по ID
func (r *GenreRepository) GetByID(ctx context.Context, id domain.ID) (*domain.Genre, error) {
	return r.getByIDInternal(ctx, nil, id)
}

// GetByName возвращает жанр по названию
func (r *GenreRepository) GetByName(ctx context.Context, name string) (*domain.Genre, error) {
	entGenre, err := r.client.Genre.Query().
		Where(genre.Name(name)).
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapGenreToDomain(entGenre), nil
}

// Delete удаляет жанр
func (r *GenreRepository) Delete(ctx context.Context, id domain.ID) error {
	err := r.client.Genre.DeleteOneID(id).Exec(ctx)
	return HandleError(err)
}

// List возвращает список жанров
func (r *GenreRepository) List(ctx context.Context, opts QueryOptions) ([]*domain.Genre, error) {
	query := r.client.Genre.Query()

	// Применяем опции пагинации
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	// Фильтрация по ID, если указаны
	if len(opts.IDs) > 0 {
		query = query.Where(genre.IDIn(opts.IDs...))
	}

	entGenres, err := query.All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return MapGenresToDomain(entGenres), nil
}

// Count возвращает количество жанров
func (r *GenreRepository) Count(ctx context.Context) (int, error) {
	count, err := r.client.Genre.Query().Count(ctx)
	if err != nil {
		return 0, HandleError(err)
	}

	return count, nil
}

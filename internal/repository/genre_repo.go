package repository

import (
	"context"

	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/ent/genre"
	"github.com/nimyab/nim2book-back/internal/domain"
)

// GenreRepository реализует domain.GenreRepository
type GenreRepository struct {
	*BaseRepository
}

// NewGenreRepository создает новый репозиторий жанров
func NewGenreRepository(client *ent.Client) *GenreRepository {
	return &GenreRepository{
		BaseRepository: NewBaseRepository(client),
	}
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
	return r.CreateTx(ctx, nil, domainGenre)
}

// CreateTx создает новый жанр внутри транзакции (если передана)
func (r *GenreRepository) CreateTx(ctx context.Context, tx *ent.Tx, domainGenre *domain.Genre) (*domain.Genre, error) {
	return DoInTxOrUse(ctx, r.client, tx, func(tx *ent.Tx) (*domain.Genre, error) {
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

// Update обновляет жанр
// Примечание: name - immutable поле, его нельзя обновить
// Этот метод загружает актуальные данные из БД
func (r *GenreRepository) Update(ctx context.Context, domainGenre *domain.Genre) (*domain.Genre, error) {
	return r.UpdateTx(ctx, nil, domainGenre)
}

// UpdateTx обновляет жанр внутри транзакции (если передана)
func (r *GenreRepository) UpdateTx(ctx context.Context, tx *ent.Tx, domainGenre *domain.Genre) (*domain.Genre, error) {
	// Так как все поля жанра immutable (кроме timestamps),
	// просто возвращаем актуальные данные
	return r.getByIDInternal(ctx, tx, domainGenre.ID)
}

// Delete удаляет жанр
func (r *GenreRepository) Delete(ctx context.Context, id domain.ID) error {
	return r.DeleteTx(ctx, nil, id)
}

// DeleteTx удаляет жанр внутри транзакции (если передана)
func (r *GenreRepository) DeleteTx(ctx context.Context, tx *ent.Tx, id domain.ID) error {
	_, err := DoInTxOrUse(ctx, r.client, tx, func(tx *ent.Tx) (struct{}, error) {
		err := tx.Genre.DeleteOneID(id).Exec(ctx)
		if err != nil {
			return struct{}{}, HandleError(err)
		}
		return struct{}{}, nil
	})
	return err
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

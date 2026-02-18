package repository

import (
	"context"
	"strings"

	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/ent/author"
	"github.com/nimyab/nim2book-back/internal/domain"
)

// AuthorRepository реализует domain.AuthorRepository
type AuthorRepository struct {
	*BaseRepository
}

// NewAuthorRepository создает новый репозиторий авторов
func NewAuthorRepository(client *ent.Client) *AuthorRepository {
	return &AuthorRepository{
		BaseRepository: NewBaseRepository(client),
	}
}

// getByIDInternal возвращает автора по ID, работает как с транзакцией, так и без неё
func (r *AuthorRepository) getByIDInternal(ctx context.Context, tx *ent.Tx, id domain.ID) (*domain.Author, error) {
	client := GetClientOrTx(r.client, tx)

	entAuthor, err := client.Author.Query().
		Where(author.ID(id)).
		Only(ctx)
	if err != nil {
		return nil, HandleError(err)
	}
	return MapAuthorToDomain(entAuthor), nil
}

// Create создает нового автора
func (r *AuthorRepository) Create(ctx context.Context, domainAuthor *domain.Author) (*domain.Author, error) {
	return r.CreateTx(ctx, nil, domainAuthor)
}

// CreateTx создает нового автора внутри транзакции (если передана)
func (r *AuthorRepository) CreateTx(ctx context.Context, tx *ent.Tx, domainAuthor *domain.Author) (*domain.Author, error) {
	return DoInTxOrUse(ctx, r.client, tx, func(tx *ent.Tx) (*domain.Author, error) {
		entAuthor, err := tx.Author.Create().
			SetName(domainAuthor.Name).
			Save(ctx)

		if err != nil {
			return nil, HandleError(err)
		}

		return r.getByIDInternal(ctx, tx, entAuthor.ID)
	})
}

// GetByID возвращает автора по ID
func (r *AuthorRepository) GetByID(ctx context.Context, id domain.ID) (*domain.Author, error) {
	return r.getByIDInternal(ctx, nil, id)
}

// GetByName возвращает автора по имени
func (r *AuthorRepository) GetByName(ctx context.Context, name string) (*domain.Author, error) {
	entAuthor, err := r.client.Author.Query().
		Where(author.Name(name)).
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapAuthorToDomain(entAuthor), nil
}

// Update обновляет автора
// Примечание: name - immutable поле, его нельзя обновить
// Этот метод загружает актуальные данные из БД
func (r *AuthorRepository) Update(ctx context.Context, domainAuthor *domain.Author) (*domain.Author, error) {
	// Так как все поля автора immutable (кроме timestamps),
	// просто возвращаем актуальные данные
	return r.GetByID(ctx, domainAuthor.ID)
}

// GetOrCreate возвращает существующего автора или создаёт нового
func (r *AuthorRepository) GetOrCreate(ctx context.Context, name string) (*domain.Author, error) {
	// Сначала пытаемся найти существующего автора
	existingAuthor, err := r.GetByName(ctx, name)
	if err == nil {
		return existingAuthor, nil
	}

	// Если автор не найден, создаём нового
	if err == ErrNotFound {
		return r.Create(ctx, &domain.Author{Name: name})
	}

	// Если произошла другая ошибка, возвращаем её
	return nil, err
}

// Delete удаляет автора
func (r *AuthorRepository) Delete(ctx context.Context, id domain.ID) error {
	return r.DeleteTx(ctx, nil, id)
}

// DeleteTx удаляет автора внутри транзакции (если передана)
func (r *AuthorRepository) DeleteTx(ctx context.Context, tx *ent.Tx, id domain.ID) error {
	_, err := DoInTxOrUse(ctx, r.client, tx, func(tx *ent.Tx) (struct{}, error) {
		err := tx.Author.DeleteOneID(id).Exec(ctx)
		if err != nil {
			return struct{}{}, HandleError(err)
		}
		return struct{}{}, nil
	})
	return err
}

// List возвращает список авторов с пагинацией
func (r *AuthorRepository) List(ctx context.Context, opts QueryOptions) ([]*domain.Author, error) {
	query := r.client.Author.Query()

	// Применяем опции пагинации
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	// Фильтрация по ID, если указаны
	if len(opts.IDs) > 0 {
		query = query.Where(author.IDIn(opts.IDs...))
	}

	entAuthors, err := query.All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return MapAuthorsToDomain(entAuthors), nil
}

// Search ищет авторов по имени
func (r *AuthorRepository) Search(ctx context.Context, searchQuery string, opts QueryOptions) ([]*domain.Author, error) {
	query := r.client.Author.Query().
		Where(author.NameContainsFold(strings.TrimSpace(searchQuery)))

	// Применяем опции пагинации
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	entAuthors, err := query.All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return MapAuthorsToDomain(entAuthors), nil
}

// Count возвращает количество авторов
func (r *AuthorRepository) Count(ctx context.Context) (int, error) {
	count, err := r.client.Author.Query().Count(ctx)
	if err != nil {
		return 0, HandleError(err)
	}

	return count, nil
}

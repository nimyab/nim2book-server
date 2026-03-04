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
	client *ent.Client
}

// NewAuthorRepository создает новый репозиторий авторов
func NewAuthorRepository(client *ent.Client) *AuthorRepository {
	return &AuthorRepository{client: client}
}

// getByIDInternal возвращает автора по ID, работает как с транзакцией, так и без неё
func (r *AuthorRepository) getByIDInternal(ctx context.Context, tx *ent.Tx, id domain.ID) (*domain.Author, error) {
	client := GetClientOrTx(r.client, tx)

	entAuthor, err := client.Author.Query().Where(author.ID(id)).Only(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return MapAuthorToDomain(entAuthor), nil
}

// Create создает нового автора
func (r *AuthorRepository) Create(ctx context.Context, domainAuthor *domain.Author) (*domain.Author, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.Author, error) {
		entAuthor, err := tx.Author.Create().SetName(domainAuthor.Name).Save(ctx)
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
	entAuthor, err := r.client.Author.Query().Where(author.Name(name)).Only(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return MapAuthorToDomain(entAuthor), nil
}

// GetOrCreate возвращает существующего автора или создаёт нового
func (r *AuthorRepository) GetOrCreate(ctx context.Context, name string) (*domain.Author, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.Author, error) {
		existEntAuthor, err := tx.Author.Query().Where(author.Name(name)).Only(ctx)
		if err != nil && !ent.IsNotFound(err) {
			return nil, HandleError(err)
		}
		if existEntAuthor != nil {
			return MapAuthorToDomain(existEntAuthor), nil
		}

		newEntAuthor, err := tx.Author.Create().SetName(name).Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		return MapAuthorToDomain(newEntAuthor), nil
	})
}

// Delete удаляет автора
func (r *AuthorRepository) Delete(ctx context.Context, id domain.ID) error {
	err := r.client.Author.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return HandleError(err)
	}
	return nil

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

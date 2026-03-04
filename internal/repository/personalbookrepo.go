package repository

import (
	"context"
	"strings"

	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/ent/author"
	"github.com/nimyab/nim2book-back/ent/genre"
	"github.com/nimyab/nim2book-back/ent/personalbook"
	"github.com/nimyab/nim2book-back/ent/personalbookchapter"
	"github.com/nimyab/nim2book-back/ent/user"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/samber/lo"
)

// PersonalBookRepository реализует domain.PersonalBookRepository
type PersonalBookRepository struct {
	client *ent.Client
}

// NewPersonalBookRepository создает новый репозиторий личных книг
func NewPersonalBookRepository(client *ent.Client) *PersonalBookRepository {
	return &PersonalBookRepository{client: client}
}

// getByIDInternal возвращает личную книгу по ID, может работать внутри транзакции
func (r *PersonalBookRepository) getByIDInternal(ctx context.Context, tx *ent.Tx, id domain.ID) (*domain.PersonalBook, error) {
	client := GetClientOrTx(r.client, tx)

	entBook, err := client.PersonalBook.Query().
		Where(personalbook.ID(id)).
		WithUser().
		WithAuthor().
		WithGenres().
		WithPersonalBookChapters().
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapPersonalBookToDomain(entBook), nil
}

// Create создает новую личную книгу
func (r *PersonalBookRepository) Create(ctx context.Context, domainBook *domain.PersonalBook) (*domain.PersonalBook, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.PersonalBook, error) {
		// Создаем личную книгу
		create := tx.PersonalBook.Create().
			SetTitle(domainBook.Title).
			SetOriginalLang(domainBook.OriginalLang).
			SetTranslatedLang(domainBook.TranslatedLang).
			SetProcessStatus(personalbook.ProcessStatus(domainBook.ProcessStatus))

		if domainBook.CoverURL != nil {
			create = create.SetCoverURL(*domainBook.CoverURL)
		}

		// Устанавливаем пользователя, если указан
		if domainBook.User != nil {
			create = create.SetUserID(domainBook.User.ID)
		}

		// Устанавливаем автора, если указан
		if domainBook.Author != nil {
			create = create.SetAuthorID(domainBook.Author.ID)
		}

		// Добавляем жанры, если указаны
		if len(domainBook.Genres) > 0 {
			create.AddGenreIDs(
				lo.Map(domainBook.Genres, func(g *domain.Genre, index int) domain.ID {
					return g.ID
				})...,
			)
		}

		entBook, err := create.Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		return r.getByIDInternal(ctx, tx, entBook.ID)
	})
}

// GetByID возвращает личную книгу по ID
func (r *PersonalBookRepository) GetByID(ctx context.Context, id domain.ID) (*domain.PersonalBook, error) {
	return r.getByIDInternal(ctx, nil, id)
}

// Update обновляет личную книгу
func (r *PersonalBookRepository) Update(ctx context.Context, domainBook *domain.PersonalBook) (*domain.PersonalBook, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.PersonalBook, error) {
		// title - immutable поле, поэтому обновляем только mutable поля
		update := tx.PersonalBook.UpdateOneID(domainBook.ID).
			SetNillableCoverURL(domainBook.CoverURL).
			SetOriginalLang(domainBook.OriginalLang).
			SetTranslatedLang(domainBook.TranslatedLang).
			SetProcessStatus(personalbook.ProcessStatus(domainBook.ProcessStatus))

		// Обновляем автора, если указан
		if domainBook.Author != nil {
			update = update.SetAuthorID(domainBook.Author.ID)
		}

		entBook, err := update.Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		// Обновляем жанры, если указаны
		if domainBook.Genres != nil {
			// Очищаем существующие жанры
			err = tx.PersonalBook.UpdateOne(entBook).ClearGenres().Exec(ctx)
			if err != nil {
				return nil, HandleError(err)
			}

			// Добавляем новые жанры
			if len(domainBook.Genres) > 0 {
				genreIDs := make([]domain.ID, len(domainBook.Genres))
				for i, g := range domainBook.Genres {
					genreIDs[i] = g.ID
				}
				err = tx.PersonalBook.UpdateOne(entBook).AddGenreIDs(genreIDs...).Exec(ctx)
				if err != nil {
					return nil, HandleError(err)
				}
			}
		}

		// Загружаем полную информацию о книге
		return r.getByIDInternal(ctx, tx, entBook.ID)
	})
}

// UpdateProcessStatus обновляет статус процесса личной книги
func (r *PersonalBookRepository) UpdateProcessStatus(ctx context.Context, id domain.ID, processStatus domain.ProcessStatus) (*domain.PersonalBook, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.PersonalBook, error) {
		// Обновляем статус процесса
		update := tx.PersonalBook.UpdateOneID(id).
			SetProcessStatus(personalbook.ProcessStatus(processStatus))

		entBook, err := update.Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		return r.getByIDInternal(ctx, tx, entBook.ID)
	})
}

// Delete удаляет личную книгу
func (r *PersonalBookRepository) Delete(ctx context.Context, id domain.ID) error {
	err := r.client.PersonalBook.DeleteOneID(id).Exec(ctx)
	return HandleError(err)
}

// ListByUserID возвращает личные книги пользователя
func (r *PersonalBookRepository) ListByUserID(ctx context.Context, userID domain.ID, opts QueryOptions) ([]*domain.PersonalBook, error) {
	query := r.client.PersonalBook.Query().
		Where(personalbook.HasUserWith(user.ID(userID))).
		WithUser().
		WithAuthor().
		WithGenres()

	// Применяем опции пагинации
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	entBooks, err := query.All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return MapPersonalBooksToDomain(entBooks), nil
}

// SearchByUserWithFilters ищет личные книги пользователя с возможностью комбинирования фильтров
func (r *PersonalBookRepository) SearchByUserWithFilters(ctx context.Context, userID domain.ID, title, authorName string, genreID *domain.ID, opts QueryOptions) ([]*domain.PersonalBook, error) {
	query := r.client.PersonalBook.Query().
		Where(personalbook.HasUserWith(user.ID(userID))).
		WithUser().
		WithAuthor().
		WithGenres()

	// Фильтрация по названию книги
	if title != "" {
		query = query.Where(personalbook.TitleContainsFold(strings.TrimSpace(title)))
	}

	// Фильтрация по имени автора
	if authorName != "" {
		query = query.Where(personalbook.HasAuthorWith(author.NameContainsFold(strings.TrimSpace(authorName))))
	}

	// Фильтрация по жанру
	if genreID != nil {
		query = query.Where(personalbook.HasGenresWith(genre.ID(*genreID)))
	}

	// Применяем опции пагинации
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	entBooks, err := query.All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return MapPersonalBooksToDomain(entBooks), nil
}

// GetByUserAndAuthorAndTitle возвращает личную книгу пользователя по имени автора и названию
func (r *PersonalBookRepository) GetByUserAndAuthorAndTitle(ctx context.Context, userID domain.ID, authorName, title string) (*domain.PersonalBook, error) {
	entBook, err := r.client.PersonalBook.Query().
		Where(
			personalbook.TitleEQ(title),
			personalbook.HasUserWith(user.ID(userID)),
			personalbook.HasAuthorWith(author.NameEQ(authorName)),
		).
		WithAuthor().
		WithGenres().
		WithUser().
		Only(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return MapPersonalBookToDomain(entBook), nil
}

// Count возвращает количество личных книг
func (r *PersonalBookRepository) Count(ctx context.Context) (int, error) {
	count, err := r.client.PersonalBook.Query().Count(ctx)
	if err != nil {
		return 0, HandleError(err)
	}

	return count, nil
}

// ============================================================================
// Методы для работы с главами личных книг
// ============================================================================

// getChapterByIDInternal возвращает главу личной книги по ID, может работать внутри транзакции
func (r *PersonalBookRepository) getChapterByIDInternal(ctx context.Context, tx *ent.Tx, id domain.ID) (*domain.PersonalBookChapter, error) {
	client := GetClientOrTx(r.client, tx)

	entChapter, err := client.PersonalBookChapter.Query().
		Where(personalbookchapter.ID(id)).
		WithPersonalBook().
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapPersonalBookChapterToDomain(entChapter), nil
}

// CreateChapter создает новую главу личной книги
func (r *PersonalBookRepository) CreateChapter(ctx context.Context, domainChapter *domain.PersonalBookChapter) (*domain.PersonalBookChapter, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.PersonalBookChapter, error) {
		create := tx.PersonalBookChapter.Create().
			SetOrder(domainChapter.Order).
			SetTitle(domainChapter.Title).
			SetTranslatedTitle(domainChapter.TranslatedTitle).
			SetContentURL(domainChapter.ContentURL)

		// Устанавливаем книгу, если указана
		if domainChapter.PersonalBook != nil {
			create = create.SetPersonalBookID(domainChapter.PersonalBook.ID)
		}

		entChapter, err := create.Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		return r.getChapterByIDInternal(ctx, tx, entChapter.ID)
	})
}

// GetChapterByID возвращает главу по ID
func (r *PersonalBookRepository) GetChapterByID(ctx context.Context, id domain.ID) (*domain.PersonalBookChapter, error) {
	return r.getChapterByIDInternal(ctx, nil, id)
}

// UpdateChapter обновляет главу
func (r *PersonalBookRepository) UpdateChapter(ctx context.Context, domainChapter *domain.PersonalBookChapter) (*domain.PersonalBookChapter, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.PersonalBookChapter, error) {
		// order - immutable поле, поэтому обновляем только mutable поля
		update := tx.PersonalBookChapter.UpdateOneID(domainChapter.ID).
			SetTitle(domainChapter.Title).
			SetTranslatedTitle(domainChapter.TranslatedTitle).
			SetContentURL(domainChapter.ContentURL)

		// Обновляем книгу, если указана
		if domainChapter.PersonalBook != nil {
			update = update.SetPersonalBookID(domainChapter.PersonalBook.ID)
		}

		entChapter, err := update.Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		return r.getChapterByIDInternal(ctx, tx, entChapter.ID)
	})
}

// DeleteChapter удаляет главу
func (r *PersonalBookRepository) DeleteChapter(ctx context.Context, id domain.ID) error {
	err := r.client.PersonalBookChapter.DeleteOneID(id).Exec(ctx)
	return HandleError(err)
}

// ListChaptersByPersonalBookID возвращает главы личной книги
func (r *PersonalBookRepository) ListChaptersByPersonalBookID(ctx context.Context, personalBookID domain.ID) ([]*domain.PersonalBookChapter, error) {
	entChapters, err := r.client.PersonalBookChapter.Query().
		Where(personalbookchapter.HasPersonalBookWith(personalbook.ID(personalBookID))).
		Order(ent.Asc(personalbookchapter.FieldOrder)).
		All(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapPersonalBookChaptersToDomain(entChapters), nil
}

func (r *PersonalBookRepository) GetChapterByPersonalBookIDAndOrder(ctx context.Context, personalBookID domain.ID, orderChapter int) (*domain.PersonalBookChapter, error) {
	entChapter, err := r.client.PersonalBookChapter.Query().Where(
		personalbookchapter.HasPersonalBookWith(personalbook.ID(personalBookID)),
		personalbookchapter.OrderEQ(orderChapter),
	).Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapPersonalBookChapterToDomain(entChapter), nil
}

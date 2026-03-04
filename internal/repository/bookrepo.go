package repository

import (
	"context"
	"strings"

	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/ent/author"
	"github.com/nimyab/nim2book-back/ent/book"
	"github.com/nimyab/nim2book-back/ent/bookchapter"
	"github.com/nimyab/nim2book-back/ent/genre"
	"github.com/nimyab/nim2book-back/internal/domain"
)

// BookRepository реализует domain.BookRepository
type BookRepository struct {
	client *ent.Client
}

// NewBookRepository создает новый репозиторий книг
func NewBookRepository(client *ent.Client) *BookRepository {
	return &BookRepository{client: client}
}

// getByIDInternal возвращает книгу по ID, может работать внутри транзакции
func (r *BookRepository) getByIDInternal(ctx context.Context, tx *ent.Tx, id domain.ID) (*domain.Book, error) {
	client := GetClientOrTx(r.client, tx)

	entBook, err := client.Book.Query().
		Where(book.ID(id)).
		WithAuthor().
		WithGenres().
		WithBookChapters().
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapBookToDomain(entBook), nil
}

// Create создает новую книгу
func (r *BookRepository) Create(ctx context.Context, domainBook *domain.Book) (*domain.Book, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.Book, error) {
		// Создаем книгу
		create := tx.Book.Create().
			SetTitle(domainBook.Title).
			SetOriginalLang(domainBook.OriginalLang).
			SetTranslatedLang(domainBook.TranslatedLang)

		// Устанавливаем обложку, если указано
		if domainBook.CoverURL != nil {
			create = create.SetCoverURL(*domainBook.CoverURL)
		}

		// Устанавливаем автора, если указан
		if domainBook.Author != nil {
			create = create.SetAuthorID(domainBook.Author.ID)
		}

		entBook, err := create.Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		// Добавляем жанры, если указаны
		if len(domainBook.Genres) > 0 {
			genreIDs := make([]domain.ID, len(domainBook.Genres))
			for i, g := range domainBook.Genres {
				genreIDs[i] = g.ID
			}
			err = tx.Book.UpdateOne(entBook).AddGenreIDs(genreIDs...).Exec(ctx)
			if err != nil {
				return nil, HandleError(err)
			}
		}

		return r.getByIDInternal(ctx, tx, entBook.ID)
	})
}

// GetByID возвращает книгу по ID
func (r *BookRepository) GetByID(ctx context.Context, id domain.ID) (*domain.Book, error) {
	return r.getByIDInternal(ctx, nil, id)
}

// GetByIDWithChapters возвращает книгу с главами
func (r *BookRepository) GetByIDWithChapters(ctx context.Context, id domain.ID) (*domain.Book, error) {
	entBook, err := r.client.Book.Query().
		Where(book.ID(id)).
		WithAuthor().
		WithGenres().
		WithBookChapters(func(q *ent.BookChapterQuery) {
			q.Order(ent.Asc("order"))
		}).
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapBookToDomain(entBook), nil
}

// Update обновляет книгу
func (r *BookRepository) Update(ctx context.Context, domainBook *domain.Book) (*domain.Book, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.Book, error) {
		update := tx.Book.UpdateOneID(domainBook.ID).
			SetNillableCoverURL(domainBook.CoverURL).
			SetOriginalLang(domainBook.OriginalLang).
			SetTranslatedLang(domainBook.TranslatedLang)

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
			err = tx.Book.UpdateOne(entBook).ClearGenres().Exec(ctx)
			if err != nil {
				return nil, HandleError(err)
			}

			// Добавляем новые жанры
			if len(domainBook.Genres) > 0 {
				genreIDs := make([]domain.ID, len(domainBook.Genres))
				for i, g := range domainBook.Genres {
					genreIDs[i] = g.ID
				}
				err = tx.Book.UpdateOne(entBook).AddGenreIDs(genreIDs...).Exec(ctx)
				if err != nil {
					return nil, HandleError(err)
				}
			}
		}

		// Загружаем полную информацию о книге
		return r.getByIDInternal(ctx, tx, entBook.ID)
	})
}

// UpdateProcessStatus обновляет статус обработки книги
func (r *BookRepository) UpdateProcessStatus(ctx context.Context, id domain.ID, processStatus domain.ProcessStatus) (*domain.Book, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.Book, error) {
		entBook, err := tx.Book.UpdateOneID(id).
			SetProcessStatus(book.ProcessStatus(processStatus)).
			Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		return r.getByIDInternal(ctx, tx, entBook.ID)
	})
}

// Delete удаляет книгу
func (r *BookRepository) Delete(ctx context.Context, id domain.ID) error {
	_, err := DoInTx(ctx, r.client, func(tx *ent.Tx) (struct{}, error) {
		err := tx.Book.DeleteOneID(id).Exec(ctx)
		if err != nil {
			return struct{}{}, HandleError(err)
		}
		return struct{}{}, nil
	})
	return err
}

// List возвращает список книг с пагинацией
func (r *BookRepository) List(ctx context.Context, opts QueryOptions) ([]*domain.Book, error) {
	query := r.client.Book.Query().
		WithAuthor().
		WithGenres().
		WithBookChapters(func(query *ent.BookChapterQuery) {
			query.Order(ent.Asc("order"))
		})

	// Применяем опции пагинации
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	// Фильтрация по ID, если указаны
	if len(opts.IDs) > 0 {
		query = query.Where(book.IDIn(opts.IDs...))
	}

	entBooks, err := query.All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return MapBooksToDomain(entBooks), nil
}

// ListByAuthorID возвращает книги автора
func (r *BookRepository) ListByAuthorID(ctx context.Context, authorID domain.ID, opts QueryOptions) ([]*domain.Book, error) {
	query := r.client.Book.Query().
		Where(book.HasAuthorWith(author.ID(authorID))).
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

	return MapBooksToDomain(entBooks), nil
}

// ListByGenreID возвращает книги по жанру
func (r *BookRepository) ListByGenreID(ctx context.Context, genreID domain.ID, opts QueryOptions) ([]*domain.Book, error) {
	query := r.client.Book.Query().
		Where(book.HasGenresWith(genre.ID(genreID))).
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

	return MapBooksToDomain(entBooks), nil
}

// Search ищет книги по названию
func (r *BookRepository) Search(ctx context.Context, searchQuery string, opts QueryOptions) ([]*domain.Book, error) {
	query := r.client.Book.Query().
		Where(book.TitleContainsFold(strings.TrimSpace(searchQuery))).
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

	return MapBooksToDomain(entBooks), nil
}

// SearchWithFilters ищет книги с возможностью комбинирования фильтров
func (r *BookRepository) SearchWithFilters(ctx context.Context, title, authorName string, genreID *domain.ID, opts QueryOptions) ([]*domain.Book, error) {
	query := r.client.Book.Query().
		WithAuthor().
		WithGenres().
		WithBookChapters(func(q *ent.BookChapterQuery) {
			q.Order(ent.Asc("order"))
		})

	// Фильтрация по названию книги
	if title != "" {
		query = query.Where(book.TitleContainsFold(strings.TrimSpace(title)))
	}

	// Фильтрация по имени автора
	if authorName != "" {
		query = query.Where(book.HasAuthorWith(author.NameContainsFold(strings.TrimSpace(authorName))))
	}

	// Фильтрация по жанру
	if genreID != nil {
		query = query.Where(book.HasGenresWith(genre.ID(*genreID)))
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

	return MapBooksToDomain(entBooks), nil
}

// GetByAuthorAndTitle возвращает книгу по имени автора и названию
func (r *BookRepository) GetByAuthorAndTitle(ctx context.Context, authorName, title string) (*domain.Book, error) {
	entBook, err := r.client.Book.Query().
		Where(
			book.TitleEQ(title),
			book.HasAuthorWith(author.NameEQ(authorName)),
		).
		WithAuthor().
		WithGenres().
		Only(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return MapBookToDomain(entBook), nil
}

// Count возвращает количество книг
func (r *BookRepository) Count(ctx context.Context) (int, error) {
	count, err := r.client.Book.Query().Count(ctx)
	if err != nil {
		return 0, HandleError(err)
	}

	return count, nil
}

// ============================================================================
// Методы для работы с главами книг
// ============================================================================

// getChapterByIDInternal возвращает главу по ID, может работать внутри транзакции
func (r *BookRepository) getChapterByIDInternal(ctx context.Context, tx *ent.Tx, id domain.ID) (*domain.BookChapter, error) {
	client := GetClientOrTx(r.client, tx)

	entChapter, err := client.BookChapter.Query().
		Where(bookchapter.ID(id)).
		WithBook().
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapBookChapterToDomain(entChapter), nil
}

// CreateChapter создает новую главу книги
func (r *BookRepository) CreateChapter(ctx context.Context, domainChapter *domain.BookChapter) (*domain.BookChapter, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.BookChapter, error) {
		create := tx.BookChapter.Create().
			SetOrder(domainChapter.Order).
			SetTitle(domainChapter.Title).
			SetTranslatedTitle(domainChapter.TranslatedTitle).
			SetContentURL(domainChapter.ContentURL)

		// Устанавливаем книгу, если указана
		if domainChapter.Book != nil {
			create = create.SetBookID(domainChapter.Book.ID)
		}

		entChapter, err := create.Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		return r.getChapterByIDInternal(ctx, tx, entChapter.ID)
	})
}

// GetChapterByID возвращает главу по ID
func (r *BookRepository) GetChapterByID(ctx context.Context, id domain.ID) (*domain.BookChapter, error) {
	return r.getChapterByIDInternal(ctx, nil, id)
}

// UpdateChapter обновляет главу
func (r *BookRepository) UpdateChapter(ctx context.Context, domainChapter *domain.BookChapter) (*domain.BookChapter, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.BookChapter, error) {
		// order - immutable поле, поэтому обновляем только mutable поля
		update := tx.BookChapter.UpdateOneID(domainChapter.ID).
			SetTitle(domainChapter.Title).
			SetTranslatedTitle(domainChapter.TranslatedTitle).
			SetContentURL(domainChapter.ContentURL)

		// Обновляем книгу, если указана
		if domainChapter.Book != nil {
			update = update.SetBookID(domainChapter.Book.ID)
		}

		entChapter, err := update.Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		return r.getChapterByIDInternal(ctx, tx, entChapter.ID)
	})
}

// DeleteChapter удаляет главу
func (r *BookRepository) DeleteChapter(ctx context.Context, id domain.ID) error {
	err := r.client.BookChapter.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return HandleError(err)
	}
	return nil
}

// ListChaptersByBookID возвращает главы книги
func (r *BookRepository) ListChaptersByBookID(ctx context.Context, bookID domain.ID, opts QueryOptions) ([]*domain.BookChapter, error) {
	query := r.client.BookChapter.Query().
		Where(bookchapter.HasBookWith(book.ID(bookID))).
		WithBook().
		Order(ent.Asc(bookchapter.FieldOrder))

	// Применяем опции пагинации
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	entChapters, err := query.All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return MapBookChaptersToDomain(entChapters), nil
}

// CountChapters возвращает количество глав
func (r *BookRepository) CountChapters(ctx context.Context) (int, error) {
	count, err := r.client.BookChapter.Query().Count(ctx)
	if err != nil {
		return 0, HandleError(err)
	}

	return count, nil
}

// GetChapterByBookIDAndOrder возвращает главу книги по ID книги и порядку главы
func (r *BookRepository) GetChapterByBookIDAndOrder(ctx context.Context, bookID domain.ID, orderChapter int) (*domain.BookChapter, error) {
	entChapter, err := r.client.BookChapter.Query().Where(
		bookchapter.HasBookWith(book.ID(bookID)),
		bookchapter.OrderEQ(orderChapter),
	).Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapBookChapterToDomain(entChapter), nil
}

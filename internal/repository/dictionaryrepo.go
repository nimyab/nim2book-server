package repository

import (
	"context"

	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/ent/dictionary"
	"github.com/nimyab/nim2book-back/ent/dictionaryexample"
	"github.com/nimyab/nim2book-back/internal/domain"
)

// DictionaryRepository реализует domain.DictionaryRepository
type DictionaryRepository struct {
	client *ent.Client
}

// NewDictionaryRepository создает новый репозиторий словаря
func NewDictionaryRepository(client *ent.Client) *DictionaryRepository {
	return &DictionaryRepository{client: client}
}

// getByIDInternal возвращает запись словаря по ID, может работать внутри транзакции
func (r *DictionaryRepository) getByIDInternal(ctx context.Context, tx *ent.Tx, id domain.ID) (*domain.DictionaryWord, error) {
	client := GetClientOrTx(r.client, tx)

	entDict, err := client.Dictionary.Query().
		Where(dictionary.ID(id)).
		WithDictionaryExamples().
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapDictionaryToDomain(entDict), nil
}

// Create создает новую запись в словаре
func (r *DictionaryRepository) Create(ctx context.Context, domainDict *domain.DictionaryWord) (*domain.DictionaryWord, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.DictionaryWord, error) {
		// Создаём примеры, если они есть
		var exampleIDs []domain.ID
		if len(domainDict.Examples) > 0 {
			for _, example := range domainDict.Examples {
				exampleBuilder := tx.DictionaryExample.Create().
					SetText(example.Text).
					SetTranslation(example.TranslatedText)

				if example.WordPositionStart != nil {
					exampleBuilder = exampleBuilder.SetTargetPositionStart(*example.WordPositionStart)
				}
				if example.WordPositionEnd != nil {
					exampleBuilder = exampleBuilder.SetTargetPositionEnd(*example.WordPositionEnd)
				}

				entExample, err := exampleBuilder.Save(ctx)
				if err != nil {
					return nil, HandleError(err)
				}

				exampleIDs = append(exampleIDs, entExample.ID)
			}
		}

		// Создаём словарную запись
		dictBuilder := tx.Dictionary.Create().
			SetText(domainDict.Text).
			SetFromLangCode(domainDict.FromLangCode).
			SetToLangCode(domainDict.ToLangCode).
			SetPartOfSpeech(domainDict.PartOfSpeech).
			SetTranslations(domainDict.Translations)

		if domainDict.Transcription != nil {
			dictBuilder = dictBuilder.SetTranscription(*domainDict.Transcription)
		}

		if len(exampleIDs) > 0 {
			dictBuilder = dictBuilder.AddDictionaryExampleIDs(exampleIDs...)
		}

		entDict, err := dictBuilder.Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		return r.getByIDInternal(ctx, tx, entDict.ID)
	})
}

// GetByID возвращает запись словаря по ID
func (r *DictionaryRepository) GetByID(ctx context.Context, id domain.ID) (*domain.DictionaryWord, error) {
	return r.getByIDInternal(ctx, nil, id)
}

// Update обновляет запись словаря
func (r *DictionaryRepository) Update(ctx context.Context, domainDict *domain.DictionaryWord) (*domain.DictionaryWord, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.DictionaryWord, error) {
		update := tx.Dictionary.UpdateOneID(domainDict.ID).
			SetText(domainDict.Text).
			SetFromLangCode(domainDict.FromLangCode).
			SetToLangCode(domainDict.ToLangCode).
			SetPartOfSpeech(domainDict.PartOfSpeech).
			SetTranslations(domainDict.Translations)

		// Обновляем транскрипцию
		if domainDict.Transcription != nil {
			update = update.SetTranscription(*domainDict.Transcription)
		} else {
			update = update.SetTranscription("")
		}

		entDict, err := update.Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		return r.getByIDInternal(ctx, tx, entDict.ID)
	})
}

// Delete удаляет запись словаря
func (r *DictionaryRepository) Delete(ctx context.Context, id domain.ID) error {
	_, err := DoInTx(ctx, r.client, func(tx *ent.Tx) (struct{}, error) {
		err := tx.Dictionary.DeleteOneID(id).Exec(ctx)
		if err != nil {
			return struct{}{}, HandleError(err)
		}
		return struct{}{}, nil
	})
	return err
}

// List возвращает список записей словаря
func (r *DictionaryRepository) List(ctx context.Context, opts QueryOptions) ([]*domain.DictionaryWord, error) {
	query := r.client.Dictionary.Query().WithDictionaryExamples()

	// Применяем опции пагинации
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	// Фильтрация по ID, если указаны
	if len(opts.IDs) > 0 {
		query = query.Where(dictionary.IDIn(opts.IDs...))
	}

	entDicts, err := query.All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return MapDictionariesToDomain(entDicts), nil
}

// Count возвращает количество записей
func (r *DictionaryRepository) Count(ctx context.Context) (int, error) {
	count, err := r.client.Dictionary.Query().Count(ctx)
	if err != nil {
		return 0, HandleError(err)
	}

	return count, nil
}

// GetDictionaryWordsByText возвращает слова по тексту и языкам
func (r *DictionaryRepository) GetDictionaryWordsByText(ctx context.Context, text, fromLang, toLang string) ([]*domain.DictionaryWord, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) ([]*domain.DictionaryWord, error) {
		entDicts, err := tx.Dictionary.Query().
			Where(
				dictionary.Text(text),
				dictionary.FromLangCode(fromLang),
				dictionary.ToLangCode(toLang),
			).
			WithDictionaryExamples().
			All(ctx)

		if err != nil {
			return nil, HandleError(err)
		}

		return MapDictionariesToDomain(entDicts), nil
	})
}

// ============================================================================
// Методы для работы с примерами словаря
// ============================================================================

// getExampleByIDInternal возвращает пример по ID внутри транзакции (если передана)
func (r *DictionaryRepository) getExampleByIDInternal(ctx context.Context, tx *ent.Tx, id domain.ID) (*domain.DictionaryExample, error) {
	client := GetClientOrTx(r.client, tx)

	entExample, err := client.DictionaryExample.Query().
		Where(dictionaryexample.ID(id)).
		WithDictionary().
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return new(MapDictionaryExampleToDomain(entExample)), nil
}

// AddExample добавляет новый пример к слову в словаре
func (r *DictionaryRepository) AddExample(ctx context.Context, dictionaryID domain.ID, example *domain.DictionaryExample) (*domain.DictionaryExample, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.DictionaryExample, error) {
		// Проверяем, существует ли словарная запись
		exists, err := tx.Dictionary.Query().Where(dictionary.ID(dictionaryID)).Exist(ctx)
		if err != nil {
			return nil, HandleError(err)
		}
		if !exists {
			return nil, ErrNotFound
		}

		// Создаем пример
		create := tx.DictionaryExample.Create().
			SetText(example.Text).
			SetTranslation(example.TranslatedText).
			SetDictionaryID(dictionaryID)

		if example.WordPositionStart != nil {
			create = create.SetTargetPositionStart(*example.WordPositionStart)
		}
		if example.WordPositionEnd != nil {
			create = create.SetTargetPositionEnd(*example.WordPositionEnd)
		}

		entExample, err := create.Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		return r.getExampleByIDInternal(ctx, tx, entExample.ID)
	})
}

// GetExampleByID возвращает пример по ID
func (r *DictionaryRepository) GetExampleByID(ctx context.Context, id domain.ID) (*domain.DictionaryExample, error) {
	return r.getExampleByIDInternal(ctx, nil, id)
}

// UpdateExample обновляет пример
func (r *DictionaryRepository) UpdateExample(ctx context.Context, example *domain.DictionaryExample) (*domain.DictionaryExample, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.DictionaryExample, error) {
		update := tx.DictionaryExample.UpdateOneID(example.ID).
			SetText(example.Text).
			SetTranslation(example.TranslatedText).
			SetNillableTargetPositionStart(example.WordPositionStart).
			SetNillableTargetPositionEnd(example.WordPositionEnd)

		_, err := update.Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		return r.getExampleByIDInternal(ctx, tx, example.ID)
	})
}

// DeleteExample удаляет пример
func (r *DictionaryRepository) DeleteExample(ctx context.Context, id domain.ID) error {
	err := r.client.DictionaryExample.DeleteOneID(id).Exec(ctx)
	if err != nil {
		return HandleError(err)
	}
	return nil
}

// ListExamplesByDictionaryID возвращает все примеры для словарной записи
func (r *DictionaryRepository) ListExamplesByDictionaryID(ctx context.Context, dictionaryID domain.ID, opts QueryOptions) ([]*domain.DictionaryExample, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) ([]*domain.DictionaryExample, error) {
		query := tx.DictionaryExample.Query().
			Where(dictionaryexample.HasDictionaryWith(dictionary.ID(dictionaryID))).
			WithDictionary()

		// Применяем опции пагинации
		if opts.Limit > 0 {
			query = query.Limit(opts.Limit)
		}
		if opts.Offset > 0 {
			query = query.Offset(opts.Offset)
		}

		entExamples, err := query.All(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		examples := make([]*domain.DictionaryExample, len(entExamples))
		for i, entExample := range entExamples {
			examples[i] = new(MapDictionaryExampleToDomain(entExample))
		}

		return examples, nil
	})
}

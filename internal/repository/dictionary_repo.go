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
	*BaseRepository
}

// NewDictionaryRepository создает новый репозиторий словаря
func NewDictionaryRepository(client *ent.Client) *DictionaryRepository {
	return &DictionaryRepository{
		BaseRepository: NewBaseRepository(client),
	}
}

// Create создает новую запись в словаре
func (r *DictionaryRepository) Create(ctx context.Context, domainDict *domain.DictionaryWord) (*domain.DictionaryWord, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.DictionaryWord, error) {
		createDictionary := tx.Dictionary.Create().
			SetText(domainDict.Text).
			SetFromLangCode(domainDict.FromLangCode).
			SetToLangCode(domainDict.ToLangCode).
			SetPartOfSpeech(domainDict.PartOfSpeech).
			SetTranslations(domainDict.Translations)

		// Устанавливаем транскрипцию, если указана
		if domainDict.Transcription != nil {
			createDictionary = createDictionary.SetTranscription(*domainDict.Transcription)
		}

		// Добавляем примеры, если они есть
		if len(domainDict.Examples) > 0 {
			for _, example := range domainDict.Examples {
				dictExample := tx.DictionaryExample.Create().
					SetText(example.Text).
					SetTranslation(example.TranslatedText)

				if example.WordPositionStart != nil {
					dictExample = dictExample.SetTargetPositionStart(*example.WordPositionStart)
				}
				if example.WordPositionEnd != nil {
					dictExample = dictExample.SetTargetPositionEnd(*example.WordPositionEnd)
				}

				entDict, err := dictExample.Save(ctx)
				if err != nil {
					return nil, HandleError(err)
				}

				createDictionary.AddDictionaryExampleIDs(entDict.ID)
			}
		}

		entDict, err := createDictionary.Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		return r.GetByID(ctx, entDict.ID)
	})
}

// GetByID возвращает запись словаря по ID
func (r *DictionaryRepository) GetByID(ctx context.Context, id domain.ID) (*domain.DictionaryWord, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.DictionaryWord, error) {
		entDict, err := tx.Dictionary.Query().
			Where(dictionary.ID(id)).
			WithDictionaryExamples().
			Only(ctx)

		if err != nil {
			return nil, HandleError(err)
		}

		return MapDictionaryToDomain(entDict), nil
	})
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

		return r.GetByID(ctx, entDict.ID)
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
	return DoInTx(ctx, r.client, func(tx *ent.Tx) ([]*domain.DictionaryWord, error) {
		query := tx.Dictionary.Query().WithDictionaryExamples()

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
	})
}

// Count возвращает количество записей
func (r *DictionaryRepository) Count(ctx context.Context) (int, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (int, error) {
		count, err := tx.Dictionary.Query().Count(ctx)
		if err != nil {
			return 0, HandleError(err)
		}

		return count, nil
	})
}

// ============================================================================
// Методы-обертки для совместимости с существующими сервисами
// ============================================================================

// CreateDictionaryWord - обертка над Create для совместимости
func (r *DictionaryRepository) CreateDictionaryWord(ctx context.Context, word *domain.DictionaryWord) (domain.ID, error) {
	created, err := r.Create(ctx, word)
	if err != nil {
		return domain.ID{}, err
	}
	return created.ID, nil
}

// CreateDictionaryExample - создает пример использования слова
// Это упрощенная версия для совместимости с lookup service
func (r *DictionaryRepository) CreateDictionaryExample(ctx context.Context, example *domain.DictionaryExample) (domain.ID, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (domain.ID, error) {
		create := tx.DictionaryExample.Create().
			SetText(example.Text).
			SetTranslation(example.TranslatedText)

		if example.WordPositionStart != nil {
			create = create.SetTargetPositionStart(*example.WordPositionStart)
		}
		if example.WordPositionEnd != nil {
			create = create.SetTargetPositionEnd(*example.WordPositionEnd)
		}

		// Связываем с dictionary, если ID указан
		if example.DictionaryID != (domain.ID{}) {
			create = create.SetDictionaryID(example.DictionaryID)
		}

		entExample, err := create.Save(ctx)
		if err != nil {
			return domain.ID{}, HandleError(err)
		}

		return entExample.ID, nil
	})
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

		return r.GetExampleByID(ctx, entExample.ID)
	})
}

// GetExampleByID возвращает пример по ID
func (r *DictionaryRepository) GetExampleByID(ctx context.Context, id domain.ID) (*domain.DictionaryExample, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.DictionaryExample, error) {
		entExample, err := tx.DictionaryExample.Query().
			Where(dictionaryexample.ID(id)).
			WithDictionary().
			Only(ctx)

		if err != nil {
			return nil, HandleError(err)
		}

		var dictionaryID domain.ID
		if entExample.Edges.Dictionary != nil {
			dictionaryID = entExample.Edges.Dictionary.ID
		}

		example := &domain.DictionaryExample{
			ID:                entExample.ID,
			Text:              entExample.Text,
			TranslatedText:    entExample.Translation,
			WordPositionStart: entExample.TargetPositionStart,
			WordPositionEnd:   entExample.TargetPositionEnd,
			DictionaryID:      dictionaryID,
		}

		return example, nil
	})
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

		return r.GetExampleByID(ctx, example.ID)
	})
}

// DeleteExample удаляет пример
func (r *DictionaryRepository) DeleteExample(ctx context.Context, id domain.ID) error {
	_, err := DoInTx(ctx, r.client, func(tx *ent.Tx) (struct{}, error) {
		err := tx.DictionaryExample.DeleteOneID(id).Exec(ctx)
		if err != nil {
			return struct{}{}, HandleError(err)
		}
		return struct{}{}, nil
	})
	return err
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
			var dictID domain.ID
			if entExample.Edges.Dictionary != nil {
				dictID = entExample.Edges.Dictionary.ID
			}

			examples[i] = &domain.DictionaryExample{
				ID:                entExample.ID,
				Text:              entExample.Text,
				TranslatedText:    entExample.Translation,
				WordPositionStart: entExample.TargetPositionStart,
				WordPositionEnd:   entExample.TargetPositionEnd,
				DictionaryID:      dictID,
			}
		}

		return examples, nil
	})
}

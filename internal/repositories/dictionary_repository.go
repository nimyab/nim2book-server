package repositories

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/models"
	"gorm.io/gorm"
)

var (
	ErrDictionaryDataNotFound      = errors.New("dictionary data not found")
	ErrDictionaryDataAlreadyExists = errors.New("dictionary data already exists")
)

type DictionaryRepository struct {
	*Repository[models.Dictionary]
	db *gorm.DB
}

func NewDictionaryRepository(db *gorm.DB) *DictionaryRepository {
	return &DictionaryRepository{
		Repository: NewRepository[models.Dictionary](db),
		db:         db,
	}
}

// GetDictionaryData retrieves dictionary data by text and language
func (r *DictionaryRepository) GetDictionaryData(ctx context.Context, text, lang string) (*models.Dictionary, error) {
	dict, err := r.Query().
		Where("text = ? AND lang = ?", text, lang).
		First(ctx)

	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, ErrDictionaryDataNotFound
		}
		return nil, err
	}

	return dict, nil
}

// CreateDictionaryData creates a new dictionary entry
func (r *DictionaryRepository) CreateDictionaryData(ctx context.Context, text, lang string, content []byte) (*models.Dictionary, error) {
	var result *models.Dictionary
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := NewRepository[models.Dictionary](tx)

		// Check if entry already exists
		exists, err := repo.Exists(ctx, map[string]interface{}{
			"text": text,
			"lang": lang,
		})
		if err != nil {
			return err
		}
		if exists {
			return ErrDictionaryDataAlreadyExists
		}

		// Create new dictionary entry
		dict := &models.Dictionary{
			Text:    text,
			Lang:    lang,
			Content: content,
		}

		if err := repo.Create(ctx, dict); err != nil {
			return err
		}

		result = dict
		return nil
	})

	if err != nil {
		if errors.Is(err, ErrDictionaryDataAlreadyExists) {
			return nil, err
		}
		return nil, err
	}

	return result, nil
}

// UpdateDictionaryData updates an existing dictionary entry
func (r *DictionaryRepository) UpdateDictionaryData(ctx context.Context, id uuid.UUID, content []byte) error {
	err := r.UpdateFields(ctx, id, map[string]interface{}{
		"content": content,
	})
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return ErrDictionaryDataNotFound
		}
		return err
	}

	return nil
}

// DeleteDictionaryData deletes a dictionary entry
func (r *DictionaryRepository) DeleteDictionaryData(ctx context.Context, id uuid.UUID) error {
	err := r.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return ErrDictionaryDataNotFound
		}
		return err
	}

	return nil
}

// FindDictionaryByLanguage retrieves all dictionary entries for a language
func (r *DictionaryRepository) FindDictionaryByLanguage(ctx context.Context, lang string, page, pageSize int) ([]*models.Dictionary, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 50
	}

	qb := r.Query().Where("lang = ?", lang)

	// Count total
	total, err := qb.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Get paginated results
	offset := (page - 1) * pageSize
	dicts, err := qb.Limit(pageSize).Offset(offset).Order("text ASC").Find(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*models.Dictionary, len(dicts))
	for i := range dicts {
		result[i] = &dicts[i]
	}

	return result, total, nil
}

// SearchDictionary searches dictionary entries by text pattern
func (r *DictionaryRepository) SearchDictionary(ctx context.Context, pattern, lang string, limit int) ([]*models.Dictionary, error) {
	if limit < 1 {
		limit = 20
	}

	dicts, err := r.Query().
		Where("lang = ? AND text ILIKE ?", lang, "%"+pattern+"%").
		Limit(limit).
		Order("text ASC").
		Find(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]*models.Dictionary, len(dicts))
	for i := range dicts {
		result[i] = &dicts[i]
	}

	return result, nil
}

// CountDictionaryByLanguage returns the number of dictionary entries for a language
func (r *DictionaryRepository) CountDictionaryByLanguage(ctx context.Context, lang string) (int64, error) {
	count, err := r.Count(ctx, map[string]interface{}{"lang": lang})
	if err != nil {
		return 0, err
	}

	return count, nil
}

// GetSupportedLanguages returns a list of unique languages in the dictionary
func (r *DictionaryRepository) GetSupportedLanguages(ctx context.Context) ([]string, error) {
	var languages []string
	result := r.db.WithContext(ctx).
		Model(&models.Dictionary{}).
		Distinct("lang").
		Pluck("lang", &languages)

	if result.Error != nil {
		return nil, result.Error
	}

	return languages, nil
}

// BulkCreateDictionary creates multiple dictionary entries in a transaction
func (r *DictionaryRepository) BulkCreateDictionary(ctx context.Context, entries []struct {
	Text    string
	Lang    string
	Content []byte
}) (int, error) {
	var created int
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := NewRepository[models.Dictionary](tx)

		for _, entry := range entries {
			// Check if exists
			exists, err := repo.Exists(ctx, map[string]interface{}{
				"text": entry.Text,
				"lang": entry.Lang,
			})
			if err != nil {
				return err
			}
			if exists {
				continue // Skip existing entries
			}

			// Create entry
			dict := &models.Dictionary{
				Text:    entry.Text,
				Lang:    entry.Lang,
				Content: entry.Content,
			}

			if err := repo.Create(ctx, dict); err != nil {
				return err
			}

			created++
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return created, nil
}

// ParseContent unmarshals the dictionary content JSON
func (r *DictionaryRepository) ParseContent(dict *models.Dictionary) (*models.DictionaryData, error) {
	var dictData models.DictionaryData
	if err := json.Unmarshal(dict.Content, &dictData); err != nil {
		return nil, err
	}
	return &dictData, nil
}

// MarshalContent marshals dictionary data to JSON bytes
func (r *DictionaryRepository) MarshalContent(dictData *models.DictionaryData) ([]byte, error) {
	return json.Marshal(dictData)
}

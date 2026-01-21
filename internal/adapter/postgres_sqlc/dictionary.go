package postgres_sqlc

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/nimyab/nim2book-back/db/sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/logger"
	"github.com/nimyab/nim2book-back/pkg/transaction"
)

var (
	ErrDictionaryWordNotFound      = errors.New("dictionary word not found")
	ErrDictionaryWordAlreadyExists = errors.New("dictionary word already exists")
)

// GetDictionaryWordByText получает слово по тексту и языковой паре
func (db *Postgres) GetDictionaryWordByText(
	ctx context.Context,
	text, fromLang, toLang, partOfSpeech string,
) (*domain.DictionaryWord, error) {
	const operation = "postgres_sqlc.GetDictionaryWordByText"

	dictData, err := db.Queries.GetDictionaryWord(ctx, sqlc.GetDictionaryWordParams{
		Text:         text,
		FromLangCode: fromLang,
		ToLangCode:   toLang,
		PartOfSpeech: partOfSpeech,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrDictionaryWordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	examplesData, err := db.Queries.GetDictionaryExamples(ctx, dictData.ID)
	if err != nil {
		logger.Error("failed to get dictionary examples", err, operation)
	}

	examples := make([]domain.DictionaryExample, len(examplesData))
	for i, ex := range examplesData {
		examples[i] = domain.DictionaryExample{
			ID:                uuidFromPgtype(ex.ID),
			Text:              ex.Text,
			TranslatedText:    ex.TranslatedText,
			WordPositionStart: int(ex.WordPositionStart),
			WordPositionEnd:   int(ex.WordPositionEnd),
			DictionaryID:      uuidFromPgtype(ex.DictionaryID),
		}
	}

	return &domain.DictionaryWord{
		ID:            uuidFromPgtype(dictData.ID),
		Text:          dictData.Text,
		FromLangCode:  dictData.FromLangCode,
		ToLangCode:    dictData.ToLangCode,
		PartOfSpeech:  dictData.PartOfSpeech,
		Translations:  dictData.Translations,
		Transcription: dictData.Transcription,
		Examples:      examples,
	}, nil
}

// GetDictionaryWordById получает слово по его ID
func (db *Postgres) GetDictionaryWordById(
	ctx context.Context,
	wordID uuid.UUID,
) (*domain.DictionaryWord, error) {
	const operation = "postgres_sqlc.GetDictionaryWordWithExamples"

	dictData, err := db.Queries.GetDictionaryWordById(ctx, uuidToPgtype(wordID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrDictionaryWordNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	examplesData, err := db.Queries.GetDictionaryExamples(ctx, dictData.ID)
	if err != nil {
		logger.Error("failed to get dictionary examples", err, operation)
	}

	examples := make([]domain.DictionaryExample, len(examplesData))
	for i, ex := range examplesData {
		examples[i] = domain.DictionaryExample{
			ID:                uuidFromPgtype(ex.ID),
			Text:              ex.Text,
			TranslatedText:    ex.TranslatedText,
			WordPositionStart: int(ex.WordPositionStart),
			WordPositionEnd:   int(ex.WordPositionEnd),
			DictionaryID:      uuidFromPgtype(ex.DictionaryID),
		}
	}

	return &domain.DictionaryWord{
		ID:            uuidFromPgtype(dictData.ID),
		Text:          dictData.Text,
		FromLangCode:  dictData.FromLangCode,
		ToLangCode:    dictData.ToLangCode,
		PartOfSpeech:  dictData.PartOfSpeech,
		Translations:  dictData.Translations,
		Transcription: dictData.Transcription,
		Examples:      examples,
	}, nil
}

// GetDictionaryWordsByText получает все варианты слова (разные части речи)
func (db *Postgres) GetDictionaryWordsByText(
	ctx context.Context,
	text, fromLang, toLang string,
) ([]domain.DictionaryWord, error) {
	const operation = "postgres_sqlc.GetDictionaryWordsByText"

	rows, err := db.Queries.GetDictionaryWordsByText(ctx, sqlc.GetDictionaryWordsByTextParams{
		Text:         text,
		FromLangCode: fromLang,
		ToLangCode:   toLang,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	words := make([]domain.DictionaryWord, len(rows))
	for i, row := range rows {
		words[i] = domain.DictionaryWord{
			ID:            uuidFromPgtype(row.ID),
			Text:          row.Text,
			FromLangCode:  row.FromLangCode,
			ToLangCode:    row.ToLangCode,
			PartOfSpeech:  row.PartOfSpeech,
			Translations:  row.Translations,
			Transcription: row.Transcription,
		}

		examplesData, err := db.Queries.GetDictionaryExamples(ctx, row.ID)
		if err != nil {
			logger.Error("failed to get dictionary examples", err, operation)
		}

		examples := make([]domain.DictionaryExample, len(examplesData))
		for j, ex := range examplesData {
			examples[j] = domain.DictionaryExample{
				ID:                uuidFromPgtype(ex.ID),
				Text:              ex.Text,
				TranslatedText:    ex.TranslatedText,
				WordPositionStart: int(ex.WordPositionStart),
				WordPositionEnd:   int(ex.WordPositionEnd),
				DictionaryID:      uuidFromPgtype(ex.DictionaryID),
			}
		}

		words[i].Examples = examples
	}

	return words, nil
}

// CreateDictionaryWord создает новое слово
func (db *Postgres) CreateDictionaryWord(
	ctx context.Context,
	word *domain.DictionaryWord,
) (uuid.UUID, error) {
	const operation = "postgres_sqlc.CreateDictionaryWord"

	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (uuid.UUID, error) {
		queries := db.Queries.WithTx(tx)

		// Проверяем существование
		exists, err := queries.DictionaryWordExists(ctx, sqlc.DictionaryWordExistsParams{
			Text:         word.Text,
			FromLangCode: word.FromLangCode,
			ToLangCode:   word.ToLangCode,
			PartOfSpeech: word.PartOfSpeech,
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("%s: %w", operation, err)
		}
		if exists {
			return uuid.Nil, ErrDictionaryWordAlreadyExists
		}

		// Создаем слово
		pgWordID, err := queries.CreateDictionaryWord(ctx, sqlc.CreateDictionaryWordParams{
			Text:          word.Text,
			FromLangCode:  word.FromLangCode,
			ToLangCode:    word.ToLangCode,
			PartOfSpeech:  word.PartOfSpeech,
			Translations:  word.Translations,
			Transcription: word.Transcription,
		})
		if err != nil {
			return uuid.Nil, fmt.Errorf("%s: %w", operation, err)
		}

		return uuidFromPgtype(pgWordID), nil
	})
}

// CreateDictionaryExample создает пример использования слова
func (db *Postgres) CreateDictionaryExample(
	ctx context.Context,
	example *domain.DictionaryExample,
) (uuid.UUID, error) {
	const operation = "postgres_sqlc.CreateDictionaryExample"

	exampleID, err := db.Queries.CreateDictionaryExample(ctx, sqlc.CreateDictionaryExampleParams{
		Text:              example.Text,
		TranslatedText:    example.TranslatedText,
		WordPositionStart: int32(example.WordPositionStart),
		WordPositionEnd:   int32(example.WordPositionEnd),
		DictionaryID:      uuidToPgtype(example.DictionaryID),
	})
	if err != nil {
		return uuid.Nil, fmt.Errorf("%s: %w", operation, err)
	}

	return uuidFromPgtype(exampleID), nil
}

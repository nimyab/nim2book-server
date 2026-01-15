package postgres_sqlc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/nimyab/nim2book-back/db/sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/transaction"
)

var (
	ErrDictionaryDataNotFound      = errors.New("dictionary data not found")
	ErrDictionaryDataAlreadyExists = errors.New("dictionary data already exists")
)

func (db *Postgres) GetDictionaryData(ctx context.Context, text, lang string) (*domain.DictionaryData, error) {
	const operation = "postgres_sqlc.GetDictionaryData"

	content, err := db.Queries.GetDictionaryData(ctx, sqlc.GetDictionaryDataParams{
		Text: text,
		Lang: lang,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrDictionaryDataNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	dictData := new(domain.DictionaryData)
	if err = json.Unmarshal(content, dictData); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return dictData, nil
}

func (db *Postgres) CreateDictionaryData(
	ctx context.Context,
	text, lang string,
	dictData *domain.DictionaryData,
) (bool, error) {
	const operation = "postgres_sqlc.CreateDictionaryData"

	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (bool, error) {
		queries := db.Queries.WithTx(tx)

		// Проверяем существует ли уже запись
		exists, err := queries.DictionaryDataExists(ctx, sqlc.DictionaryDataExistsParams{
			Text: text,
			Lang: lang,
		})
		if err != nil {
			return false, fmt.Errorf("%s: %w", operation, err)
		}
		if exists {
			return false, ErrDictionaryDataAlreadyExists
		}

		data, err := json.Marshal(dictData)
		if err != nil {
			return false, fmt.Errorf("%s: %w", operation, err)
		}

		// Создаем запись
		err = queries.CreateDictionaryData(ctx, sqlc.CreateDictionaryDataParams{
			Text:    text,
			Lang:    lang,
			Content: data,
		})
		if err != nil {
			return false, fmt.Errorf("%s: %w", operation, err)
		}

		if err = tx.Commit(ctx); err != nil {
			return false, fmt.Errorf("%s: %w", operation, err)
		}

		return true, nil
	})
}

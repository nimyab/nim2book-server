package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/nimyab/nim2book-back/internal/domain"
)

var (
	ErrDictionaryDataNotFound      = errors.New("dictionary data not found")
	ErrDictionaryDataAlreadyExists = errors.New("dictionary data already exists")
)

func (db *Postgres) GetDictionaryData(ctx context.Context, text, lang string) (*domain.DictionaryData, error) {
	const operation = "postgres.GetDictionaryData"

	sql := `select content from dictionary where text = $1 and lang = $2`

	var content []byte
	err := db.Pool.QueryRow(ctx, sql, text, lang).Scan(&content)
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
	const operation = "postgres.CreateDictionaryData"

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("%s: %w", operation, err)
	}
	defer tx.Rollback(ctx)

	sql := `select exists(select id from dictionary where text = $1 and lang = $2);`
	var exists bool
	err = tx.QueryRow(ctx, sql, text, lang).Scan(&exists)
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
	sql = `insert into dictionary (text, lang, content) values ($1, $2, $3)`
	_, err = tx.Exec(ctx, sql, text, lang, data)
	if err != nil {
		return false, fmt.Errorf("%s: %w", operation, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("%s: %w", operation, err)
	}

	return true, nil
}

package transaction

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func Tx(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func TxWithData[T any](ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) (T, error)) (T, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return *new(T), err
	}
	defer tx.Rollback(ctx)

	result, err := fn(tx)
	if err != nil {
		return *new(T), err
	}

	if err := tx.Commit(ctx); err != nil {
		return *new(T), err
	}

	return result, nil
}

package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	PostgresURL string
}

type Postgres struct {
	Pool *pgxpool.Pool
}

func New(ctx context.Context, cfg *Config) (*Postgres, error) {
	const operation = "postgres.New"

	pool, err := pgxpool.New(ctx, cfg.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("%s: unable to ping database: %w", operation, err)
	}

	db := &Postgres{
		Pool: pool,
	}
	return db, nil
}

func (db *Postgres) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

package postgres_sqlc

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nimyab/nim2book-back/db/sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
)

type Config struct {
	PostgresURL string
}

type Postgres struct {
	Pool    *pgxpool.Pool
	Queries *sqlc.Queries
}

func New(ctx context.Context, cfg *Config) (*Postgres, error) {
	const operation = "postgres_sqlc.New"

	pool, err := pgxpool.New(ctx, cfg.PostgresURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("%s: unable to ping database: %w", operation, err)
	}

	db := &Postgres{
		Pool:    pool,
		Queries: sqlc.New(pool),
	}
	return db, nil
}

func (db *Postgres) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}

// Helper functions для конвертации типов
func uuidToPgtype(id domain.Id) pgtype.UUID {
	return pgtype.UUID{
		Bytes: id,
		Valid: true,
	}
}

func uuidFromPgtype(pgu pgtype.UUID) domain.Id {
	return domain.Id(pgu.Bytes)
}

func stringToPgtype(s string) pgtype.Text {
	return pgtype.Text{
		String: s,
		Valid:  true,
	}
}

func textToPgtype(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{
		String: *s,
		Valid:  true,
	}
}

func pgtypeToText(pt pgtype.Text) *string {
	if !pt.Valid {
		return nil
	}
	return &pt.String
}

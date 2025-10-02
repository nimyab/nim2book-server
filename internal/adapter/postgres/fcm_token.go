package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/nimyab/nim2book-back/internal/domain"
)

var (
	ErrFcmTokenAlreadyAdd = errors.New("token already add")
)

func (db *Postgres) GetFcmTokensByUserId(ctx context.Context, userId domain.Id) ([]domain.FcmToken, error) {
	const operation = "postgres.GetFcmTokensByUserId"

	sql := `select * from fcm_tokens where user_id = $1`
	rows, err := db.Pool.Query(ctx, sql, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer rows.Close()

	tokens, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.FcmToken])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return tokens, nil
}

func (db *Postgres) AddFcmToken(ctx context.Context, data *domain.FcmToken) (*domain.FcmToken, error) {
	const operation = "postgres.AddFcmToken"

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer tx.Rollback(ctx)

	sql := `select * from fcm_tokens where token = $1`
	err = tx.QueryRow(ctx, sql, data.Token).Scan()
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	if err == nil {
		return nil, ErrFcmTokenAlreadyAdd
	}

	sql = `insert into fcm_tokens (token, user_id) values (@token, @userId) returning token, user_id, create_at`
	args := pgx.NamedArgs{
		"token":  data.Token,
		"userId": data.UserId,
	}
	var token domain.FcmToken
	err = tx.QueryRow(ctx, sql, args).Scan(&token.Token, &token.UserId, &token.CreateAt)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &token, nil
}

func (db *Postgres) DeleteFcmToken(ctx context.Context, token string, userId domain.Id) error {
	const operation = "postgres.DeleteFcmToken"

	sql := `delete from fcm_tokens where token = $1 and user_id = $2`
	_, err := db.Pool.Exec(ctx, sql, token, userId)
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	return nil
}

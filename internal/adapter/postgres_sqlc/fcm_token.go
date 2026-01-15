package postgres_sqlc

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/nimyab/nim2book-back/db/sqlc"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/transaction"
)

var (
	ErrFcmTokenAlreadyAdd = errors.New("token already add")
)

func (db *Postgres) GetFcmTokensByUserId(ctx context.Context, userId domain.Id) ([]domain.FcmToken, error) {
	const operation = "postgres_sqlc.GetFcmTokensByUserId"

	tokens, err := db.Queries.GetFcmTokensByUserId(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	result := make([]domain.FcmToken, len(tokens))
	for i, token := range tokens {
		result[i] = fcmTokenFromSqlc(token)
	}

	return result, nil
}

func (db *Postgres) AddFcmToken(ctx context.Context, data *domain.FcmToken) (*domain.FcmToken, error) {
	const operation = "postgres_sqlc.AddFcmToken"

	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (*domain.FcmToken, error) {
		queries := db.Queries.WithTx(tx)

		// Проверяем существует ли уже токен
		_, err := queries.GetFcmTokenByToken(ctx, data.Token)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}
		if err == nil {
			return nil, ErrFcmTokenAlreadyAdd
		}

		// Добавляем токен
		token, err := queries.AddFcmToken(ctx, sqlc.AddFcmTokenParams{
			Token:  data.Token,
			UserID: data.UserId,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		result := fcmTokenFromSqlc(token)
		return &result, nil
	})
}

func (db *Postgres) DeleteFcmToken(ctx context.Context, token string, userId domain.Id) error {
	const operation = "postgres_sqlc.DeleteFcmToken"

	err := db.Queries.DeleteFcmToken(ctx, sqlc.DeleteFcmTokenParams{
		Token:  token,
		UserID: userId,
	})
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	return nil
}

// Конвертирует sqlc.FcmToken в domain.FcmToken
func fcmTokenFromSqlc(token sqlc.FcmToken) domain.FcmToken {
	return domain.FcmToken{
		Token:    token.Token,
		UserId:   token.UserID,
		CreateAt: token.CreateAt.Time,
	}
}

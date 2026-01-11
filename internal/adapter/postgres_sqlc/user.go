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
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

func (db *Postgres) CreateUserByEmailAndPassword(ctx context.Context, data *domain.EmailPasswordAccount) (*domain.User, error) {
	const operation = "postgres_sqlc.CreateUserByEmailAndPassword"

	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (*domain.User, error) {
		queries := db.Queries.WithTx(tx)

		// Проверяем существует ли уже аккаунт с таким email
		_, err := queries.GetEmailPasswordAccountByEmail(ctx, data.Email)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}
		if err == nil {
			return nil, ErrUserAlreadyExists
		}

		// Создаем email_password_account
		epaId, err := queries.CreateEmailPasswordAccount(ctx, sqlc.CreateEmailPasswordAccountParams{
			Email:        data.Email,
			PasswordHash: data.PasswordHash,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		data.Id = epaId

		// Создаем user
		userRow, err := queries.CreateUserByEmailPasswordAccountId(ctx, epaId)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		user := &domain.User{
			Id:                   userRow.ID,
			IsAdmin:              userRow.IsAdmin,
			IsVIP:                userRow.IsVip,
			EmailPasswordAccount: data,
		}

		if len(userRow.Metadata) > 0 {
			if err := json.Unmarshal(userRow.Metadata, &user.Metadata); err != nil {
				return nil, fmt.Errorf("%s: failed to unmarshal metadata: %w", operation, err)
			}
		}

		return user, nil
	})
}

func (db *Postgres) CreateUserByGoogle(ctx context.Context, data *domain.GoogleAccount) (*domain.User, error) {
	const operation = "postgres_sqlc.CreateUserByGoogle"

	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (*domain.User, error) {
		queries := db.Queries.WithTx(tx)

		// Проверяем существует ли уже user с таким Google Sub
		_, err := queries.GetUser(ctx, sqlc.GetUserParams{
			GoogleSub: &data.Sub,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}
		if err == nil {
			return nil, ErrUserAlreadyExists
		}

		// Создаем google_account
		err = queries.CreateGoogleAccount(ctx, sqlc.CreateGoogleAccountParams{
			Sub:           data.Sub,
			Email:         data.Email,
			EmailVerified: data.EmailVerified,
			Name:          data.Name,
			Picture:       data.Picture,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		// Создаем user
		userRow, err := queries.CreateUserByGoogleSub(ctx, &data.Sub)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		user := &domain.User{
			Id:            userRow.ID,
			IsAdmin:       userRow.IsAdmin,
			IsVIP:         userRow.IsVip,
			GoogleAccount: data,
		}

		if len(userRow.Metadata) > 0 {
			if err := json.Unmarshal(userRow.Metadata, &user.Metadata); err != nil {
				return nil, fmt.Errorf("%s: failed to unmarshal metadata: %w", operation, err)
			}
		}

		return user, nil
	})
}

func (db *Postgres) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	const operation = "postgres_sqlc.GetUserByEmail"

	row, err := db.Queries.GetUser(ctx, sqlc.GetUserParams{
		Email: &email,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return userRowToUser(row)
}

func (db *Postgres) GetUserByGoogleSub(ctx context.Context, sub string) (*domain.User, error) {
	const operation = "postgres_sqlc.GetUserByGoogleSub"

	row, err := db.Queries.GetUser(ctx, sqlc.GetUserParams{
		GoogleSub: &sub,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return userRowToUser(row)
}

func (db *Postgres) GetUserById(ctx context.Context, userId domain.Id) (*domain.User, error) {
	const operation = "postgres_sqlc.GetUserById"

	row, err := db.Queries.GetUser(ctx, sqlc.GetUserParams{
		UserID: userId,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return userRowToUser(row)
}

func (db *Postgres) GetUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	const operation = "postgres_sqlc.GetUser"

	var params sqlc.GetUserParams

	// Определяем по какому полю искать
	switch {
	case user.Id != domain.Id{}:
		params.UserID = user.Id
	case user.GoogleAccount != nil && user.GoogleAccount.Sub != "":
		params.GoogleSub = &user.GoogleAccount.Sub
	case user.GoogleAccount != nil && user.GoogleAccount.Email != "":
		params.GoogleEmail = &user.GoogleAccount.Email
	case user.EmailPasswordAccount != nil && user.EmailPasswordAccount.Id != domain.Id{}:
		params.EmailPasswordAccountID = user.EmailPasswordAccount.Id
	case user.EmailPasswordAccount != nil && user.EmailPasswordAccount.Email != "":
		params.Email = &user.EmailPasswordAccount.Email
	default:
		return nil, fmt.Errorf("%s: no valid identifier provided", operation)
	}

	row, err := db.Queries.GetUser(ctx, params)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return userRowToUser(row)
}

func (db *Postgres) UpdateMetadata(ctx context.Context, newMetadata domain.JsonB, userId domain.Id) (*domain.User, error) {
	const operation = "postgres_sqlc.UpdateMetadata"

	metadataBytes, err := json.Marshal(newMetadata)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal metadata: %w", operation, err)
	}

	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (*domain.User, error) {
		queries := db.Queries.WithTx(tx)

		// Получаем текущего пользователя
		user, err := db.GetUserById(ctx, userId)
		if err != nil {
			return nil, err
		}

		// Обновляем метаданные
		err = queries.UpdateUserMetadata(ctx, sqlc.UpdateUserMetadataParams{
			Metadata: metadataBytes,
			ID:       userId,
		})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		user.Metadata = newMetadata
		return user, nil
	})
}

// userRowToUser конвертирует результат SQL запроса в domain.User
func userRowToUser(row sqlc.GetUserRow) (*domain.User, error) {
	user := &domain.User{
		Id:      row.User.ID,
		IsAdmin: row.User.IsAdmin,
		IsVIP:   row.User.IsVip,
	}

	if len(row.User.Metadata) > 0 {
		if err := json.Unmarshal(row.User.Metadata, &user.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Если есть Google Account
	if row.GoogleAccount.Sub != "" {
		user.GoogleAccount = &domain.GoogleAccount{
			Sub:           row.GoogleAccount.Sub,
			Email:         row.GoogleAccount.Email,
			EmailVerified: row.GoogleAccount.EmailVerified,
			Name:          row.GoogleAccount.Name,
			Picture:       row.GoogleAccount.Picture,
		}
	}

	// Если есть Email/Password Account
	if row.EmailPasswordAccount.ID.String() != "" {
		user.EmailPasswordAccount = &domain.EmailPasswordAccount{
			Id:           row.EmailPasswordAccount.ID,
			Email:        row.EmailPasswordAccount.Email,
			PasswordHash: row.EmailPasswordAccount.PasswordHash,
		}
	}

	return user, nil
}

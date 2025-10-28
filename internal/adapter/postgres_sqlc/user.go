package postgres_sqlc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/transaction"
	"github.com/nimyab/nim2book-back/sqlc"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

// userRow - общий интерфейс для всех GetUser* строк
type userRow struct {
	ID                  pgtype.UUID
	IsAdmin             bool
	IsVip               pgtype.Bool
	Metadata            []byte
	GoogleEmail         pgtype.Text
	GoogleEmailVerified pgtype.Bool
	GoogleName          pgtype.Text
	GooglePicture       pgtype.Text
	GoogleSub           pgtype.Text
	EmailPasswordID     pgtype.UUID
	EmailPasswordEmail  pgtype.Text
	PasswordHash        pgtype.Text
}

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

		data.Id = uuidFromPgtype(epaId)

		// Создаем user
		userRow, err := queries.CreateUserByEmailPasswordAccountId(ctx, epaId)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		user := &domain.User{
			Id:                   uuidFromPgtype(userRow.ID),
			IsAdmin:              userRow.IsAdmin,
			IsVIP:                userRow.IsVip.Bool,
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
			GoogleSub: stringToPgtype(data.Sub),
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
			Picture:       textToPgtype(data.Picture),
		})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		// Создаем user
		userRow, err := queries.CreateUserByGoogleSub(ctx, stringToPgtype(data.Sub))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		user := &domain.User{
			Id:            uuidFromPgtype(userRow.ID),
			IsAdmin:       userRow.IsAdmin,
			IsVIP:         userRow.IsVip.Bool,
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
		Email: stringToPgtype(email),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return userRowToUser(userRow{
		ID:                  row.ID,
		IsAdmin:             row.IsAdmin,
		IsVip:               row.IsVip,
		Metadata:            row.Metadata,
		GoogleEmail:         row.GoogleEmail,
		GoogleEmailVerified: row.GoogleEmailVerified,
		GoogleName:          row.GoogleName,
		GooglePicture:       row.GooglePicture,
		GoogleSub:           row.GoogleSub,
		EmailPasswordID:     row.EmailPasswordID,
		EmailPasswordEmail:  row.EmailPasswordEmail,
		PasswordHash:        row.PasswordHash,
	})
}

func (db *Postgres) GetUserByGoogleSub(ctx context.Context, sub string) (*domain.User, error) {
	const operation = "postgres_sqlc.GetUserByGoogleSub"

	row, err := db.Queries.GetUser(ctx, sqlc.GetUserParams{
		GoogleSub: stringToPgtype(sub),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return userRowToUser(userRow{
		ID:                  row.ID,
		IsAdmin:             row.IsAdmin,
		IsVip:               row.IsVip,
		Metadata:            row.Metadata,
		GoogleEmail:         row.GoogleEmail,
		GoogleEmailVerified: row.GoogleEmailVerified,
		GoogleName:          row.GoogleName,
		GooglePicture:       row.GooglePicture,
		GoogleSub:           row.GoogleSub,
		EmailPasswordID:     row.EmailPasswordID,
		EmailPasswordEmail:  row.EmailPasswordEmail,
		PasswordHash:        row.PasswordHash,
	})
}

func (db *Postgres) GetUserById(ctx context.Context, userId domain.Id) (*domain.User, error) {
	const operation = "postgres_sqlc.GetUserById"

	row, err := db.Queries.GetUser(ctx, sqlc.GetUserParams{
		UserID: uuidToPgtype(userId),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return userRowToUser(userRow{
		ID:                  row.ID,
		IsAdmin:             row.IsAdmin,
		IsVip:               row.IsVip,
		Metadata:            row.Metadata,
		GoogleEmail:         row.GoogleEmail,
		GoogleEmailVerified: row.GoogleEmailVerified,
		GoogleName:          row.GoogleName,
		GooglePicture:       row.GooglePicture,
		GoogleSub:           row.GoogleSub,
		EmailPasswordID:     row.EmailPasswordID,
		EmailPasswordEmail:  row.EmailPasswordEmail,
		PasswordHash:        row.PasswordHash,
	})
}

func (db *Postgres) GetUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	const operation = "postgres_sqlc.GetUser"

	var params sqlc.GetUserParams

	// Определяем по какому полю искать
	switch {
	case user.Id != domain.Id{}:
		params.UserID = uuidToPgtype(user.Id)
	case user.GoogleAccount != nil && user.GoogleAccount.Sub != "":
		params.GoogleSub = stringToPgtype(user.GoogleAccount.Sub)
	case user.GoogleAccount != nil && user.GoogleAccount.Email != "":
		params.GoogleEmail = stringToPgtype(user.GoogleAccount.Email)
	case user.EmailPasswordAccount != nil && user.EmailPasswordAccount.Id != domain.Id{}:
		params.EmailPasswordAccountID = uuidToPgtype(user.EmailPasswordAccount.Id)
	case user.EmailPasswordAccount != nil && user.EmailPasswordAccount.Email != "":
		params.Email = stringToPgtype(user.EmailPasswordAccount.Email)
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

	return userRowToUser(userRow{
		ID:                  row.ID,
		IsAdmin:             row.IsAdmin,
		IsVip:               row.IsVip,
		Metadata:            row.Metadata,
		GoogleEmail:         row.GoogleEmail,
		GoogleEmailVerified: row.GoogleEmailVerified,
		GoogleName:          row.GoogleName,
		GooglePicture:       row.GooglePicture,
		GoogleSub:           row.GoogleSub,
		EmailPasswordID:     row.EmailPasswordID,
		EmailPasswordEmail:  row.EmailPasswordEmail,
		PasswordHash:        row.PasswordHash,
	})
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
			ID:       uuidToPgtype(userId),
		})
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		user.Metadata = newMetadata
		return user, nil
	})
}

// userRowToUser конвертирует результат SQL запроса в domain.User
func userRowToUser(row userRow) (*domain.User, error) {
	user := &domain.User{
		Id:      uuidFromPgtype(row.ID),
		IsAdmin: row.IsAdmin,
		IsVIP:   row.IsVip.Bool,
	}

	if len(row.Metadata) > 0 {
		if err := json.Unmarshal(row.Metadata, &user.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Если есть Google Account
	if row.GoogleSub.Valid {
		user.GoogleAccount = &domain.GoogleAccount{
			Sub:           row.GoogleSub.String,
			Email:         row.GoogleEmail.String,
			EmailVerified: row.GoogleEmailVerified.Bool,
			Name:          row.GoogleName.String,
			Picture:       pgtypeToText(row.GooglePicture),
		}
	}

	// Если есть Email/Password Account
	if row.EmailPasswordID.Valid {
		user.EmailPasswordAccount = &domain.EmailPasswordAccount{
			Id:           uuidFromPgtype(row.EmailPasswordID),
			Email:        row.EmailPasswordEmail.String,
			PasswordHash: row.PasswordHash.String,
		}
	}

	return user, nil
}

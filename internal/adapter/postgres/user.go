package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/transaction"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

func (db *Postgres) CreateUserByEmailAndPassword(ctx context.Context, data *domain.EmailPasswordAccount) (*domain.User, error) {
	const operation = "postgres.CreateUserByEmailAndPassword"

	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (*domain.User, error) {
		sql := `select id from email_password_accounts where email = $1;`
		err := tx.QueryRow(ctx, sql, data.Email).Scan(&data.Id)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}
		if err == nil {
			return nil, ErrUserAlreadyExists
		}

		// создаем email_password_account
		sql = `insert into email_password_accounts (email, password_hash) values (@email, @passwordHash) returning id;`
		args := pgx.NamedArgs{
			"email":        data.Email,
			"passwordHash": data.PasswordHash,
		}
		err = tx.QueryRow(ctx, sql, args).Scan(&data.Id)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		// создаем user
		sql = `insert into users (email_password_account_id) values (@epaId) returning id, is_admin, is_vip, metadata;`
		args = pgx.NamedArgs{"epaId": data.Id}
		user := &domain.User{}
		var metadata []byte
		err = tx.QueryRow(ctx, sql, args).Scan(&user.Id, &user.IsAdmin, &user.IsVIP, &metadata)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}
		user.EmailPasswordAccount = data

		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &user.Metadata); err != nil {
				return nil, fmt.Errorf("%s: failed to unmarshal metadata: %w", operation, err)
			}
		}

		return user, nil
	})
}

func (db *Postgres) CreateUserByGoogle(ctx context.Context, data *domain.GoogleAccount) (*domain.User, error) {
	const operation = "postgres.CreateUserByGoogle"

	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (*domain.User, error) {
		_, err := db.getUser(ctx, tx, &domain.User{
			GoogleAccount: &domain.GoogleAccount{
				Sub: data.Sub,
			},
		})
		if err != nil && !errors.Is(err, ErrUserNotFound) {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}
		if err == nil {
			return nil, ErrUserAlreadyExists
		}

		// создаем google_account
		sql := `insert into google_accounts (sub, email, email_verified, name, picture) 
			values (@sub, @email, @emailVerified, @name, @picture);`
		args := pgx.NamedArgs{
			"sub":           data.Sub,
			"email":         data.Email,
			"emailVerified": data.EmailVerified,
			"name":          data.Name,
			"picture":       data.Picture,
		}
		_, err = tx.Exec(ctx, sql, args)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		// создаем user
		sql = `insert into users (google_account_sub) values (@googleSub) returning id, is_admin, is_vip, metadata;`
		args = pgx.NamedArgs{"googleSub": data.Sub}
		user := &domain.User{}
		var metadata []byte
		err = tx.QueryRow(ctx, sql, args).Scan(&user.Id, &user.IsAdmin, &user.IsVIP, &metadata)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}
		user.GoogleAccount = data

		if len(metadata) > 0 {
			if err := json.Unmarshal(metadata, &user.Metadata); err != nil {
				return nil, fmt.Errorf("%s: failed to unmarshal metadata: %w", operation, err)
			}
		}

		return user, nil
	})
}

func (db *Postgres) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (*domain.User, error) {
		return db.getUser(ctx, tx, &domain.User{
			EmailPasswordAccount: &domain.EmailPasswordAccount{Email: email},
		})
	})
}

func (db *Postgres) GetUserByGoogleSub(ctx context.Context, sub string) (*domain.User, error) {
	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (*domain.User, error) {
		return db.getUser(ctx, tx, &domain.User{
			GoogleAccount: &domain.GoogleAccount{Sub: sub},
		})
	})
}

func (db *Postgres) GetUserById(ctx context.Context, userId domain.Id) (*domain.User, error) {
	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (*domain.User, error) {
		return db.getUser(ctx, tx, &domain.User{Id: userId})
	})
}

func (db *Postgres) GetUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (*domain.User, error) {
		return db.getUser(ctx, tx, user)
	})
}

// GetUser возвращает пользователя по любому из заполненных идентификаторов:
// - user.Id (ID пользователя)
// - user.GoogleAccount.Sub (Google Sub)
// - user.GoogleAccount.Email (Google Email)
// - user.EmailPasswordAccount.Id (ID email/password аккаунта)
// - user.EmailPasswordAccount.Email (Email from email/password account)
func (db *Postgres) getUser(ctx context.Context, tx pgx.Tx, user *domain.User) (*domain.User, error) {
	const operation = "postgres.GetUser"

	var (
		sqlCondition string
		args         pgx.NamedArgs
		b            strings.Builder
	)

	_, err := b.WriteString(`
			select u.id, u.is_admin, u.is_vip, u.metadata,
			g.email, g.email_verified, g.name, g.picture, g.sub,
			e.id, e.email, e.password_hash
			from users as u 
			left join google_accounts as g on u.google_account_sub = g.sub
			left join email_password_accounts as e on u.email_password_account_id = e.id

			`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	// Определяем по какому полю искать
	switch {
	case user.Id != domain.Id{}:
		// Поиск по ID пользователя
		sqlCondition = `where u.id = @userId;`
		args = pgx.NamedArgs{"userId": user.Id}

	case user.GoogleAccount != nil && user.GoogleAccount.Sub != "":
		// Поиск по Google Sub
		sqlCondition = `where u.google_account_sub = @googleSub;`
		args = pgx.NamedArgs{"googleSub": user.GoogleAccount.Sub}

	case user.GoogleAccount != nil && user.GoogleAccount.Email != "":
		// Поиск по Google Email
		sqlCondition = `where g.email = @googleEmail;`
		args = pgx.NamedArgs{"googleEmail": user.GoogleAccount.Email}

	case user.EmailPasswordAccount != nil && user.EmailPasswordAccount.Id != domain.Id{}:
		// Поиск по Email/Password Account ID
		sqlCondition = `where u.email_password_account_id = @epaId;`
		args = pgx.NamedArgs{"epaId": user.EmailPasswordAccount.Id}

	case user.EmailPasswordAccount != nil && user.EmailPasswordAccount.Email != "":
		// Поиск по Email from email/password account
		sqlCondition = `where e.email = @email;`
		args = pgx.NamedArgs{"email": user.EmailPasswordAccount.Email}

	default:
		return nil, fmt.Errorf("%s: no valid identifier provided", operation)
	}

	if _, err := b.WriteString(sqlCondition); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	var (
		sql = b.String()

		id             domain.Id
		isAdmin, isVIP bool
		metadata       []byte

		gSub, gEmail, gName *string
		gEmailVerified      *bool
		gPicture            *string

		eId                   *domain.Id
		eEmail, ePasswordHash *string
	)

	err = tx.QueryRow(ctx, sql, args).Scan(
		&id, &isAdmin, &isVIP, &metadata,
		&gEmail, &gEmailVerified, &gName, &gPicture, &gSub,
		&eId, &eEmail, &ePasswordHash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	resultUser := &domain.User{
		Id:      id,
		IsAdmin: isAdmin,
		IsVIP:   isVIP,
	}

	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &resultUser.Metadata); err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal metadata: %w", operation, err)
		}
	}

	if gSub != nil {
		resultUser.GoogleAccount = &domain.GoogleAccount{
			Sub:           *gSub,
			Email:         *gEmail,
			EmailVerified: *gEmailVerified,
			Name:          *gName,
			Picture:       gPicture,
		}
	}
	if eId != nil {
		resultUser.EmailPasswordAccount = &domain.EmailPasswordAccount{
			Id:           *eId,
			Email:        *eEmail,
			PasswordHash: *ePasswordHash,
		}
	}

	return resultUser, nil
}

func (db *Postgres) UpdateMetadata(ctx context.Context, newMetadata domain.JsonB, userId domain.Id) (*domain.User, error) {
	const operation = "postgres.UpdateMetadata"

	metadataBytes, err := json.Marshal(newMetadata)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal metadata: %w", operation, err)
	}

	return transaction.TxWithData(ctx, db.Pool, func(tx pgx.Tx) (*domain.User, error) {
		user, err := db.getUser(ctx, tx, &domain.User{Id: userId})
		if err != nil {
			return nil, err
		}

		sql := `update users set metadata = @metadata where id = @userId;`
		args := pgx.NamedArgs{
			"metadata": metadataBytes,
			"userId":   userId,
		}
		_, err = tx.Exec(ctx, sql, args)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", operation, err)
		}

		user.Metadata = newMetadata
		return user, nil
	})
}

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
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

func (db *Postgres) CreateUserByEmailAndPassword(ctx context.Context, data *domain.EmailPasswordAccount) (*domain.User, error) {
	const operation = "postgres.CreateUserByEmailAndPassword"

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer tx.Rollback(ctx)

	sql := `select id from email_password_accounts where email = $1;`
	err = tx.QueryRow(ctx, sql, data.Email).Scan(&data.Id)
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

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return user, nil
}

func (db *Postgres) CreateUserByGoogle(ctx context.Context, data *domain.GoogleAccount) (*domain.User, error) {
	const operation = "postgres.CreateUserByGoogle"

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer tx.Rollback(ctx)

	sql := `select sub from google_accounts where sub = $1;`
	var sub string
	err = db.Pool.QueryRow(ctx, sql, data.Sub).Scan(&sub)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	if err == nil {
		return nil, ErrUserAlreadyExists
	}

	// создаем google_account
	sql = `insert into google_accounts (sub, email, email_verified, name, picture) 
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

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return user, nil
}

func (db *Postgres) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	const operation = "postgres.GetUserByEmail"

	sql := `select u.id, u.is_admin, u.is_vip, u.metadata,
       				e.id, e.email, e.password_hash
			from users as u
			left join email_password_accounts as e on e.id = u.email_password_account_id
			where e.email = @email;`
	args := pgx.NamedArgs{
		"email": email,
	}

	var (
		id             domain.Id
		isAdmin, isVIP bool
		metadata       []byte

		eId                   *domain.Id
		eEmail, ePasswordHash *string
	)
	err := db.Pool.QueryRow(ctx, sql, args).Scan(
		&id, &isAdmin, &isVIP, &metadata,
		&eId, &eEmail, &ePasswordHash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	user := &domain.User{
		Id:      id,
		IsAdmin: isAdmin,
		IsVIP:   isVIP,
		EmailPasswordAccount: &domain.EmailPasswordAccount{
			Id:           *eId,
			Email:        *eEmail,
			PasswordHash: *ePasswordHash,
		},
		GoogleAccount: nil,
	}

	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &user.Metadata); err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal metadata: %w", operation, err)
		}
	}

	return user, nil
}

func (db *Postgres) GetUserByGoogleSub(ctx context.Context, sub string) (*domain.User, error) {
	const operation = "postgres.GetUserByGoogleSub"

	sql := `select u.id, u.is_admin, u.is_vip, u.metadata,
       				g.sub, g.email, g.email_verified, g.name, g.picture
			from users as u
			left join google_accounts as g on g.sub = u.google_account_sub
			where g.sub = @sub;`
	args := pgx.NamedArgs{"sub": sub}

	var (
		id             domain.Id
		isAdmin, isVIP bool
		metadata       []byte

		gSub, gEmail, gName *string
		gEmailVerified      *bool
		gPicture            *string
	)
	err := db.Pool.QueryRow(ctx, sql, args).Scan(
		&id, &isAdmin, &isVIP, &metadata,
		&gSub, &gEmail, &gEmailVerified, &gName, &gPicture,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	user := &domain.User{
		Id:      id,
		IsAdmin: isAdmin,
		IsVIP:   isVIP,
		GoogleAccount: &domain.GoogleAccount{
			Sub:           *gSub,
			Email:         *gEmail,
			Name:          *gName,
			EmailVerified: *gEmailVerified,
			Picture:       gPicture,
		},
		EmailPasswordAccount: nil,
	}

	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &user.Metadata); err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal metadata: %w", operation, err)
		}
	}

	return user, nil
}

func (db *Postgres) GetUserById(ctx context.Context, userId domain.Id) (*domain.User, error) {
	const operation = "postgres.GetUserById"

	sql := `select u.id, u.is_admin, u.is_vip, u.metadata,
       			g.email, g.email_verified, g.name, g.picture, g.sub,
       			e.id, e.email, e.password_hash
			from users as u 
			left join google_accounts as g on u.google_account_sub = g.sub
			left join email_password_accounts as e on u.email_password_account_id = e.id
         	where u.id = @userId;`
	args := pgx.NamedArgs{"userId": userId}

	var (
		id             domain.Id
		isAdmin, isVIP bool
		metadata       []byte // JSONB хранится как []byte

		gSub, gEmail, gName *string
		gEmailVerified      *bool
		gPicture            *string

		eId                   *domain.Id
		eEmail, ePasswordHash *string
	)

	err := db.Pool.QueryRow(ctx, sql, args).Scan(
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

	user := &domain.User{
		Id:      id,
		IsAdmin: isAdmin,
		IsVIP:   isVIP,
	}

	if len(metadata) > 0 {
		if err := json.Unmarshal(metadata, &user.Metadata); err != nil {
			return nil, fmt.Errorf("%s: failed to unmarshal metadata: %w", operation, err)
		}
	}

	if gSub != nil {
		user.GoogleAccount = &domain.GoogleAccount{
			Sub:           *gSub,
			Email:         *gEmail,
			EmailVerified: *gEmailVerified,
			Name:          *gName,
			Picture:       gPicture,
		}
	}
	if eId != nil {
		user.EmailPasswordAccount = &domain.EmailPasswordAccount{
			Id:           *eId,
			Email:        *eEmail,
			PasswordHash: *ePasswordHash,
		}
	}

	return user, nil
}

func (db *Postgres) UpdateMetadata(ctx context.Context, newMetadata map[string]any, userId domain.Id) (*domain.User, error) {
	const operation = "postgres.UpdateMetadata"

	metadataBytes, err := json.Marshal(newMetadata)
	if err != nil {
		return nil, fmt.Errorf("%s: failed to marshal metadata: %w", operation, err)
	}

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer tx.Rollback(ctx)

	sql := `select exists(select id from users where id = $1);`
	var exists bool
	err = tx.QueryRow(ctx, sql, userId).Scan(&exists)
	if !exists {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	sql = `update users set metadata = @metadata where id = @userId;`
	args := pgx.NamedArgs{
		"metadata": metadataBytes,
		"userId":   userId,
	}
	_, err = tx.Exec(ctx, sql, args)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	updatedUser, err := db.GetUserById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return updatedUser, nil
}

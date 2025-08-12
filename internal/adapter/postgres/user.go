package postgres

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/nimyab/nim2book-back/internal/domain"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

func (db *Postgres) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	const operation = "postgres.CreateUser"

	sql := `select * from users where email = $1`

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, sql, user.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	rows.Close()

	sql = `insert into users (email, password_hash, is_admin) values (@email, @passwordHash, @isAdmin) returning id`
	args := pgx.NamedArgs{
		"email":        user.Email,
		"passwordHash": user.PasswordHash,
		"isAdmin":      user.IsAdmin,
	}

	var id domain.Id
	err = tx.QueryRow(ctx, sql, args).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	user.Id = id
	return user, nil
}

func (db *Postgres) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	const operation = "postgres.GetUserByEmail"

	sql := `select * from users where email = $1`

	user := new(domain.User)
	err := db.Pool.QueryRow(ctx, sql, email).Scan(&user.Id, &user.Email, &user.PasswordHash, &user.IsAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return user, nil
}

func (db *Postgres) GetUserById(ctx context.Context, userId domain.Id) (*domain.User, error) {
	const operation = "postgres.GetUserById"

	sql := `select * from users where id = $1`

	user := new(domain.User)
	err := db.Pool.QueryRow(ctx, sql, userId).Scan(&user.Id, &user.Email, &user.PasswordHash, &user.IsAdmin)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return user, nil
}

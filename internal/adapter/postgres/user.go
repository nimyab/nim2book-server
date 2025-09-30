package postgres

import (
	"context"
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
	sql = `insert into users (email_password_account_id) values (@epaId) returning id, is_admin, is_vip;`
	args = pgx.NamedArgs{"epaId": data.Id}
	user := &domain.User{}
	err = tx.QueryRow(ctx, sql, args).Scan(&user.Id, &user.IsAdmin, &user.IsVIP)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	user.EmailPasswordAccount = data

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
	sql = `insert into users (google_account_sub) values (@googleSub) returning id, is_admin, is_vip`
	args = pgx.NamedArgs{"googleSub": data.Sub}
	user := &domain.User{}
	err = tx.QueryRow(ctx, sql, args).Scan(&user.Id, &user.IsAdmin, &user.IsVIP)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}
	user.GoogleAccount = data

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return user, nil
}

func (db *Postgres) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	const operation = "postgres.GetUserByEmail"

	sql := `select u.id, u.is_admin, u.is_vip,
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

		eId                   *domain.Id
		eEmail, ePasswordHash *string
	)
	err := db.Pool.QueryRow(ctx, sql, args).Scan(
		&id, &isAdmin, &isVIP,
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

	return user, nil
}

func (db *Postgres) GetUserByGoogleSub(ctx context.Context, sub string) (*domain.User, error) {
	const operation = "postgres.GetUserByGoogleSub"

	sql := `select u.id, u.is_admin, u.is_vip,
       				g.sub, g.email, g.email_verified, g.name, g.picture
			from users as u
			left join google_accounts as g on g.sub = u.google_account_sub
			where g.sub = @sub;`
	args := pgx.NamedArgs{"sub": sub}

	var (
		id             domain.Id
		isAdmin, isVIP bool

		gSub, gEmail, gName *string
		gEmailVerified      *bool
		gPicture            *string
	)
	err := db.Pool.QueryRow(ctx, sql, args).Scan(
		&id, &isAdmin, &isVIP,
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

	return user, nil
}

func (db *Postgres) GetUserById(ctx context.Context, userId domain.Id) (*domain.User, error) {
	const operation = "postgres.GetUserById"

	sql := `select u.id, u.is_admin, u.is_vip,
       			g.email, g.emailVerified, g.name, g.picture, g.sub,
       			e.id, e.email, e.password_hash
			from users as u 
			left join google_accounts as g on u.google_account_sub = g.sub
			left join email_password_accounts as e on u.email_password_account_id = e.id
         	where u.id = @userId;`
	args := pgx.NamedArgs{"userId": userId}

	var (
		id             domain.Id
		isAdmin, isVIP bool

		gSub, gEmail, gName *string
		gEmailVerified      *bool
		gPicture            *string

		eId                   *domain.Id
		eEmail, ePasswordHash *string
	)

	err := db.Pool.QueryRow(ctx, sql, args).Scan(
		&id, &isAdmin, &isVIP,
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

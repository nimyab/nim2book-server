package repository

import (
	"context"

	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/ent/basicaccount"
	"github.com/nimyab/nim2book-back/ent/googleaccount"
	"github.com/nimyab/nim2book-back/ent/user"
	"github.com/nimyab/nim2book-back/internal/domain"
)

// UserRepository реализует domain.UserRepository
type UserRepository struct {
	client *ent.Client
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(client *ent.Client) *UserRepository {
	return &UserRepository{client: client}
}

// getByIDInternal возвращает пользователя по ID, может работать внутри транзакции
func (r *UserRepository) getByIDInternal(ctx context.Context, tx *ent.Tx, id domain.ID) (*domain.User, error) {
	client := GetClientOrTx(r.client, tx)

	entUser, err := client.User.Query().
		Where(user.ID(id)).
		WithGoogleAccount().
		WithBasicAccount().
		WithPersonalBooks(func(q *ent.PersonalBookQuery) {
			q.WithAuthor().WithGenres()
		}).
		WithFcmTokens().
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapUserToDomain(entUser), nil
}

// GetByID возвращает пользователя по ID
func (r *UserRepository) GetByID(ctx context.Context, id domain.ID) (*domain.User, error) {
	return r.getByIDInternal(ctx, nil, id)
}

// GetByGoogleAccountID возвращает пользователя по Google Account ID
func (r *UserRepository) GetByGoogleAccountID(ctx context.Context, googleAccountID domain.ID) (*domain.User, error) {
	entUser, err := r.client.User.Query().
		Where(user.HasGoogleAccountWith(googleaccount.ID(googleAccountID))).
		WithGoogleAccount().
		WithBasicAccount().
		WithPersonalBooks(func(q *ent.PersonalBookQuery) {
			q.WithAuthor().WithGenres()
		}).
		WithFcmTokens().
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapUserToDomain(entUser), nil
}

// GetByBasicAccountEmail возвращает пользователя по email базового аккаунта
func (r *UserRepository) GetByBasicAccountEmail(ctx context.Context, email string) (*domain.User, error) {
	entUser, err := r.client.User.Query().
		Where(user.HasBasicAccountWith(basicaccount.Email(email))).
		WithGoogleAccount().
		WithBasicAccount().
		WithPersonalBooks(func(q *ent.PersonalBookQuery) {
			q.WithAuthor().WithGenres()
		}).
		WithFcmTokens().
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapUserToDomain(entUser), nil
}

// Update обновляет пользователя
func (r *UserRepository) Update(ctx context.Context, domainUser *domain.User) (*domain.User, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.User, error) {
		update := tx.User.UpdateOneID(domainUser.ID).
			SetIsVip(domainUser.IsVIP).
			SetIsAdmin(domainUser.IsAdmin).
			SetMetadata(domainUser.Metadata)

		entUser, err := update.Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		// Загружаем полную информацию о пользователе
		return r.getByIDInternal(ctx, tx, entUser.ID)
	})
}

// Delete удаляет пользователя
func (r *UserRepository) Delete(ctx context.Context, id domain.ID) error {
	err := r.client.User.DeleteOneID(id).Exec(ctx)
	return HandleError(err)
}

// List возвращает список пользователей с пагинацией
func (r *UserRepository) List(ctx context.Context, opts QueryOptions) ([]*domain.User, error) {
	query := r.client.User.Query().
		WithGoogleAccount().
		WithBasicAccount()

	// Применяем опции пагинации
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	// Фильтрация по ID, если указаны
	if len(opts.IDs) > 0 {
		query = query.Where(user.IDIn(opts.IDs...))
	}

	entUsers, err := query.All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return MapUsersToDomain(entUsers), nil
}

// Count возвращает количество пользователей
func (r *UserRepository) Count(ctx context.Context) (int, error) {
	count, err := r.client.User.Query().Count(ctx)
	if err != nil {
		return 0, HandleError(err)
	}

	return count, nil
}

// CreateWithGoogleAccount создает пользователя с Google аккаунтом атомарно
func (r *UserRepository) CreateWithGoogleAccount(ctx context.Context, domainUser *domain.User, googleAccount *domain.GoogleAccount) (*domain.User, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.User, error) {
		// Создаем Google аккаунт
		entGoogleAccount, err := tx.GoogleAccount.Create().
			SetSub(googleAccount.Sub).
			SetEmail(googleAccount.Email).
			SetEmailVerified(googleAccount.EmailVerified).
			SetName(googleAccount.Name).
			SetPicture(googleAccount.Picture).
			Save(ctx)

		if err != nil {
			return nil, HandleError(err)
		}

		// Создаем пользователя и связываем с Google аккаунтом
		entUser, err := tx.User.Create().
			SetIsVip(domainUser.IsVIP).
			SetIsAdmin(domainUser.IsAdmin).
			SetMetadata(domainUser.Metadata).
			SetGoogleAccount(entGoogleAccount).
			Save(ctx)

		if err != nil {
			return nil, HandleError(err)
		}

		// Загружаем полную информацию
		return r.getByIDInternal(ctx, tx, entUser.ID)
	})
}

// CreateWithBasicAccount создает пользователя с базовым аккаунтом атомарно
func (r *UserRepository) CreateWithBasicAccount(ctx context.Context, domainUser *domain.User, basicAccount *domain.BasicAccount) (*domain.User, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.User, error) {
		// Создаем базовый аккаунт
		entBasicAccount, err := tx.BasicAccount.Create().
			SetEmail(basicAccount.Email).
			SetPasswordHash(basicAccount.PasswordHash).
			SetIsVerified(basicAccount.IsVerified).
			SetVerifyLink(basicAccount.VerifyLink).
			Save(ctx)

		if err != nil {
			return nil, HandleError(err)
		}

		// Создаем пользователя и связываем с базовым аккаунтом
		entUser, err := tx.User.Create().
			SetIsVip(domainUser.IsVIP).
			SetIsAdmin(domainUser.IsAdmin).
			SetMetadata(domainUser.Metadata).
			SetBasicAccount(entBasicAccount).
			Save(ctx)

		if err != nil {
			return nil, HandleError(err)
		}

		// Загружаем полную информацию
		return r.getByIDInternal(ctx, tx, entUser.ID)
	})
}

// AttachGoogleAccount присоединяет Google аккаунт к существующему пользователю
func (r *UserRepository) AttachGoogleAccount(ctx context.Context, userID domain.ID, googleAccount *domain.GoogleAccount) error {
	_, err := DoInTx(ctx, r.client, func(tx *ent.Tx) (struct{}, error) {
		// Создаем Google аккаунт
		entGoogleAccount, err := tx.GoogleAccount.Create().
			SetSub(googleAccount.Sub).
			SetEmail(googleAccount.Email).
			SetEmailVerified(googleAccount.EmailVerified).
			SetName(googleAccount.Name).
			SetPicture(googleAccount.Picture).
			Save(ctx)

		if err != nil {
			return struct{}{}, HandleError(err)
		}

		// Связываем с пользователем
		err = tx.User.UpdateOneID(userID).
			SetGoogleAccount(entGoogleAccount).
			Exec(ctx)

		if err != nil {
			return struct{}{}, HandleError(err)
		}

		return struct{}{}, nil
	})
	return err
}

// GetGoogleAccountBySub возвращает Google аккаунт по sub (Google ID)
func (r *UserRepository) GetGoogleAccountBySub(ctx context.Context, sub string) (*domain.GoogleAccount, error) {
	entAccount, err := r.client.GoogleAccount.Query().
		Where(googleaccount.Sub(sub)).
		WithUser().
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapGoogleAccountToDomain(entAccount), nil
}

// GetGoogleAccountByEmail возвращает Google аккаунт по email
func (r *UserRepository) GetGoogleAccountByEmail(ctx context.Context, email string) (*domain.GoogleAccount, error) {
	entAccount, err := r.client.GoogleAccount.Query().
		Where(googleaccount.Email(email)).
		WithUser().
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapGoogleAccountToDomain(entAccount), nil
}

// GetBasicAccountByEmail возвращает базовый аккаунт по email (с паролем для проверки)
func (r *UserRepository) GetBasicAccountByEmail(ctx context.Context, email string) (*domain.BasicAccount, error) {
	entAccount, err := r.client.BasicAccount.Query().
		Where(basicaccount.Email(email)).
		WithUser().
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapBasicAccountToDomain(entAccount), nil
}

// GetBasicAccountByVerifyLink возвращает базовый аккаунт по токену верификации
func (r *UserRepository) GetBasicAccountByVerifyLink(ctx context.Context, verifyLink string) (*domain.BasicAccount, error) {
	entAccount, err := r.client.BasicAccount.Query().
		Where(basicaccount.VerifyLink(verifyLink)).
		WithUser().
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapBasicAccountToDomain(entAccount), nil
}

// DeleteGoogleAccount удаляет Google аккаунт пользователя
func (r *UserRepository) DeleteGoogleAccount(ctx context.Context, googleAccountID domain.ID) error {
	err := r.client.GoogleAccount.DeleteOneID(googleAccountID).Exec(ctx)
	return HandleError(err)
}

// DeleteBasicAccount удаляет базовый аккаунт пользователя
func (r *UserRepository) DeleteBasicAccount(ctx context.Context, basicAccountID domain.ID) error {
	err := r.client.BasicAccount.DeleteOneID(basicAccountID).Exec(ctx)
	return HandleError(err)

}

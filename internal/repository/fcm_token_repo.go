package repository

import (
	"context"

	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/ent/fcmtoken"
	"github.com/nimyab/nim2book-back/ent/user"
	"github.com/nimyab/nim2book-back/internal/domain"
)

// FcmTokenRepository реализует domain.FcmTokenRepository
type FcmTokenRepository struct {
	*BaseRepository
}

// NewFcmTokenRepository создает новый репозиторий FCM токенов
func NewFcmTokenRepository(client *ent.Client) *FcmTokenRepository {
	return &FcmTokenRepository{
		BaseRepository: NewBaseRepository(client),
	}
}

// getByIDInternal возвращает FCM токен по ID, может работать внутри транзакции
func (r *FcmTokenRepository) getByIDInternal(ctx context.Context, tx *ent.Tx, id domain.ID) (*domain.FcmToken, error) {
	if tx != nil {
		entToken, err := tx.FcmToken.Query().
			Where(fcmtoken.ID(id)).
			WithUser().
			Only(ctx)

		if err != nil {
			return nil, HandleError(err)
		}

		return MapFcmTokenToDomain(entToken), nil
	}

	entToken, err := r.client.FcmToken.Query().
		Where(fcmtoken.ID(id)).
		WithUser().
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapFcmTokenToDomain(entToken), nil
}

// Create создает новый FCM токен
func (r *FcmTokenRepository) Create(ctx context.Context, domainToken *domain.FcmToken) (*domain.FcmToken, error) {
	return DoInTx(ctx, r.client, func(tx *ent.Tx) (*domain.FcmToken, error) {
		create := tx.FcmToken.Create().
			SetToken(domainToken.Token)

		// Устанавливаем пользователя, если указан
		if domainToken.User != nil {
			create = create.SetUserID(domainToken.User.ID)
		}

		entToken, err := create.Save(ctx)
		if err != nil {
			return nil, HandleError(err)
		}

		return r.getByIDInternal(ctx, tx, entToken.ID)
	})
}

// GetByID возвращает FCM токен по ID
func (r *FcmTokenRepository) GetByID(ctx context.Context, id domain.ID) (*domain.FcmToken, error) {
	return r.getByIDInternal(ctx, nil, id)
}

// GetByToken возвращает FCM токен по значению токена
func (r *FcmTokenRepository) GetByToken(ctx context.Context, token string) (*domain.FcmToken, error) {
	entToken, err := r.client.FcmToken.Query().
		Where(fcmtoken.Token(token)).
		WithUser().
		Only(ctx)

	if err != nil {
		return nil, HandleError(err)
	}

	return MapFcmTokenToDomain(entToken), nil
}

// ListByUserID возвращает все FCM токены пользователя
func (r *FcmTokenRepository) ListByUserID(ctx context.Context, userID domain.ID, opts QueryOptions) ([]*domain.FcmToken, error) {
	query := r.client.FcmToken.Query().
		Where(fcmtoken.HasUserWith(user.ID(userID))).
		WithUser()

	// Применяем опции пагинации
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	entTokens, err := query.All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return MapFcmTokensToDomain(entTokens), nil
}

// Delete удаляет FCM токен
func (r *FcmTokenRepository) Delete(ctx context.Context, id domain.ID) error {
	_, err := DoInTx(ctx, r.client, func(tx *ent.Tx) (struct{}, error) {
		err := tx.FcmToken.DeleteOneID(id).Exec(ctx)
		if err != nil {
			return struct{}{}, HandleError(err)
		}
		return struct{}{}, nil
	})
	return err
}

// DeleteByToken удаляет FCM токен по значению токена
func (r *FcmTokenRepository) DeleteByToken(ctx context.Context, token string) error {
	_, err := DoInTx(ctx, r.client, func(tx *ent.Tx) (struct{}, error) {
		_, err := tx.FcmToken.Delete().
			Where(fcmtoken.Token(token)).
			Exec(ctx)
		if err != nil {
			return struct{}{}, HandleError(err)
		}
		return struct{}{}, nil
	})
	return err
}

// DeleteByUserID удаляет все FCM токены пользователя
func (r *FcmTokenRepository) DeleteByUserID(ctx context.Context, userID domain.ID) error {
	_, err := DoInTx(ctx, r.client, func(tx *ent.Tx) (struct{}, error) {
		_, err := tx.FcmToken.Delete().
			Where(fcmtoken.HasUserWith(user.ID(userID))).
			Exec(ctx)
		if err != nil {
			return struct{}{}, HandleError(err)
		}
		return struct{}{}, nil
	})
	return err
}

// List возвращает список FCM токенов с пагинацией
func (r *FcmTokenRepository) List(ctx context.Context, opts QueryOptions) ([]*domain.FcmToken, error) {
	query := r.client.FcmToken.Query().WithUser()

	// Применяем опции пагинации
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}

	// Фильтрация по ID, если указаны
	if len(opts.IDs) > 0 {
		query = query.Where(fcmtoken.IDIn(opts.IDs...))
	}

	entTokens, err := query.All(ctx)
	if err != nil {
		return nil, HandleError(err)
	}

	return MapFcmTokensToDomain(entTokens), nil
}

// Count возвращает количество FCM токенов
func (r *FcmTokenRepository) Count(ctx context.Context) (int, error) {
	count, err := r.client.FcmToken.Query().Count(ctx)
	if err != nil {
		return 0, HandleError(err)
	}

	return count, nil
}

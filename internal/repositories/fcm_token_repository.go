package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/models"
	"gorm.io/gorm"
)

var (
	ErrFcmTokenAlreadyAdd = errors.New("token already add")
	ErrFcmTokenNotFound   = errors.New("fcm token not found")
)

type FcmTokenRepository struct {
	*Repository[models.FcmToken]
	db *gorm.DB
}

func NewFcmTokenRepository(db *gorm.DB) *FcmTokenRepository {
	return &FcmTokenRepository{
		Repository: NewRepository[models.FcmToken](db),
		db:         db,
	}
}

// GetFcmTokensByUserId retrieves all FCM tokens for a user
func (r *FcmTokenRepository) GetFcmTokensByUserId(ctx context.Context, userId uuid.UUID) ([]*models.FcmToken, error) {
	tokens, err := r.Query().
		Where("user_id = ?", userId).
		Order("create_at DESC").
		Find(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]*models.FcmToken, len(tokens))
	for i := range tokens {
		result[i] = &tokens[i]
	}

	return result, nil
}

// AddFcmToken adds a new FCM token for a user
func (r *FcmTokenRepository) AddFcmToken(ctx context.Context, token string, userId uuid.UUID) (*models.FcmToken, error) {
	var result *models.FcmToken
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repo := NewRepository[models.FcmToken](tx)

		// Check if token already exists
		exists, err := repo.Exists(ctx, map[string]interface{}{"token": token})
		if err != nil {
			return err
		}
		if exists {
			return ErrFcmTokenAlreadyAdd
		}

		// Add token
		fcmToken := &models.FcmToken{
			Token:  token,
			UserID: userId,
		}

		if err := repo.Create(ctx, fcmToken); err != nil {
			return err
		}

		result = fcmToken
		return nil
	})

	return result, err
}

// DeleteFcmToken deletes an FCM token
func (r *FcmTokenRepository) DeleteFcmToken(ctx context.Context, token string, userId uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("token = ? AND user_id = ?", token, userId).
		Delete(&models.FcmToken{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return ErrFcmTokenNotFound
	}

	return nil
}

// DeleteFcmTokensByUserId deletes all FCM tokens for a user
func (r *FcmTokenRepository) DeleteFcmTokensByUserId(ctx context.Context, userId uuid.UUID) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ?", userId).
		Delete(&models.FcmToken{})

	if result.Error != nil {
		return result.Error
	}

	return nil
}

// GetFcmTokenByToken retrieves a specific FCM token
func (r *FcmTokenRepository) GetFcmTokenByToken(ctx context.Context, token string) (*models.FcmToken, error) {
	fcmToken, err := r.Query().
		Where("token = ?", token).
		Preload("User").
		First(ctx)

	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, ErrFcmTokenNotFound
		}
		return nil, err
	}

	return fcmToken, nil
}

// CountFcmTokensByUserId returns the number of FCM tokens for a user
func (r *FcmTokenRepository) CountFcmTokensByUserId(ctx context.Context, userId uuid.UUID) (int64, error) {
	count, err := r.Count(ctx, map[string]interface{}{"user_id": userId})
	if err != nil {
		return 0, err
	}

	return count, nil
}

// CleanupExpiredTokens deletes tokens older than the specified duration
func (r *FcmTokenRepository) CleanupExpiredTokens(ctx context.Context, olderThan time.Duration) (int64, error) {
	expiryDate := time.Now().UTC().Add(-olderThan)

	result := r.db.WithContext(ctx).
		Where("create_at < ?", expiryDate).
		Delete(&models.FcmToken{})

	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// TokenExists checks if a token exists
func (r *FcmTokenRepository) TokenExists(ctx context.Context, token string) (bool, error) {
	return r.Exists(ctx, map[string]interface{}{"token": token})
}

package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/models"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)

type UserRepository struct {
	*Repository[models.User]
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		Repository: NewRepository[models.User](db),
		db:         db,
	}
}

// CreateUserByEmailAndPassword creates a new user with email/password authentication
func (r *UserRepository) CreateUserByEmailAndPassword(ctx context.Context, email, passwordHash string) (*models.User, error) {
	var result *models.User
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		emailRepo := NewRepository[models.EmailPasswordAccount](tx)

		// Check if account with this email already exists
		exists, err := emailRepo.Exists(ctx, map[string]interface{}{"email": email})
		if err != nil {
			return err
		}
		if exists {
			return ErrUserAlreadyExists
		}

		// Create email_password_account
		account := &models.EmailPasswordAccount{
			Email:        email,
			PasswordHash: passwordHash,
		}

		if err := emailRepo.Create(ctx, account); err != nil {
			return err
		}

		// Create user
		user := &models.User{
			IsAdmin:                false,
			IsVip:                  false,
			Metadata:               make(models.JSONB),
			EmailPasswordAccountID: &account.ID,
		}

		userRepo := NewRepository[models.User](tx)
		if err := userRepo.Create(ctx, user); err != nil {
			return err
		}

		result = user
		return nil
	})

	return result, err
}

// CreateUserByGoogle creates a new user with Google authentication
func (r *UserRepository) CreateUserByGoogle(ctx context.Context, googleAccount *models.GoogleAccount) (*models.User, error) {
	var result *models.User
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		userRepo := NewRepository[models.User](tx)

		// Check if user with this Google Sub already exists
		exists, err := userRepo.Exists(ctx, map[string]interface{}{"google_account_sub": googleAccount.Sub})
		if err != nil {
			return err
		}
		if exists {
			return ErrUserAlreadyExists
		}

		// Create google_account
		googleRepo := NewRepository[models.GoogleAccount](tx)
		if err := googleRepo.Create(ctx, googleAccount); err != nil {
			return err
		}

		// Create user
		user := &models.User{
			IsAdmin:          false,
			IsVip:            false,
			Metadata:         make(models.JSONB),
			GoogleAccountSub: &googleAccount.Sub,
		}

		if err := userRepo.Create(ctx, user); err != nil {
			return err
		}

		result = user
		return nil
	})

	return result, err
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).
		Preload("GoogleAccount").
		Preload("EmailPasswordAccount").
		Joins("LEFT JOIN email_password_accounts ON users.email_password_account_id = email_password_accounts.id").
		Where("email_password_accounts.email = ?", email).
		First(&user)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}

	return &user, nil
}

// GetUserByGoogleSub retrieves a user by Google sub
func (r *UserRepository) GetUserByGoogleSub(ctx context.Context, sub string) (*models.User, error) {
	user, err := r.WithPreload("GoogleAccount", "EmailPasswordAccount").
		Query().
		Where("google_account_sub = ?", sub).
		First(ctx)

	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return user, nil
}

// GetUserById retrieves a user by ID
func (r *UserRepository) GetUserById(ctx context.Context, userId uuid.UUID) (*models.User, error) {
	var user models.User
	result := r.db.WithContext(ctx).
		Preload("GoogleAccount").
		Preload("EmailPasswordAccount").
		First(&user, "id = ?", userId)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}

	return &user, nil
}

// UpdateMetadata updates user metadata
func (r *UserRepository) UpdateMetadata(ctx context.Context, userId uuid.UUID, metadata models.JSONB) error {
	err := r.UpdateFields(ctx, userId, map[string]interface{}{
		"metadata": metadata,
	})
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

// UpdateUserRole updates user admin/VIP status
func (r *UserRepository) UpdateUserRole(ctx context.Context, userId uuid.UUID, isAdmin, isVIP bool) error {
	err := r.UpdateFields(ctx, userId, map[string]interface{}{
		"is_admin": isAdmin,
		"is_vip":   isVIP,
	})
	if err != nil {
		if errors.Is(err, ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}
	return nil
}

// GetUsersByRole retrieves users by role with pagination
func (r *UserRepository) GetUsersByRole(ctx context.Context, isAdmin, isVIP *bool, page, pageSize int) ([]*models.User, int64, error) {
	qb := r.Query().
		Preload("GoogleAccount").
		Preload("EmailPasswordAccount")

	if isAdmin != nil {
		qb = qb.Where("is_admin = ?", *isAdmin)
	}
	if isVIP != nil {
		qb = qb.Where("is_vip = ?", *isVIP)
	}

	// Count total
	total, err := qb.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	users, err := qb.Limit(pageSize).Offset(offset).Find(ctx)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*models.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}

	return result, total, nil
}

// GetAllAdmins retrieves all admin users
func (r *UserRepository) GetAllAdmins(ctx context.Context) ([]*models.User, error) {
	users, err := r.Query().
		Preload("GoogleAccount").
		Preload("EmailPasswordAccount").
		Where("is_admin = ?", true).
		Find(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]*models.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}

	return result, nil
}

// GetAllVIPs retrieves all VIP users
func (r *UserRepository) GetAllVIPs(ctx context.Context) ([]*models.User, error) {
	users, err := r.Query().
		Preload("GoogleAccount").
		Preload("EmailPasswordAccount").
		Where("is_vip = ?", true).
		Find(ctx)

	if err != nil {
		return nil, err
	}

	result := make([]*models.User, len(users))
	for i := range users {
		result[i] = &users[i]
	}

	return result, nil
}

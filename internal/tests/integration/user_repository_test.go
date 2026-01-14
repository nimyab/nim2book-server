package integration

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestUserRepository_CreateUserByEmailAndPassword(t *testing.T) {
	cleanupDatabase(t)

	email := "test@example.com"
	password := "password123"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := userRepo.CreateUserByEmailAndPassword(context.Background(), email, string(passwordHash))
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.False(t, user.IsAdmin)
	assert.False(t, user.IsVip)
	assert.NotNil(t, user.EmailPasswordAccountID)
}

func TestUserRepository_CreateUserByGoogle(t *testing.T) {
	cleanupDatabase(t)

	googleAccount := &models.GoogleAccount{
		Sub:           "google-sub-123",
		Email:         "google@example.com",
		EmailVerified: true,
		Name:          "Google User",
		Picture:       nil,
	}

	user, err := userRepo.CreateUserByGoogle(context.Background(), googleAccount)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.NotNil(t, user.GoogleAccountSub)
	assert.Equal(t, "google-sub-123", *user.GoogleAccountSub)
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	cleanupDatabase(t)

	email := "find@example.com"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	createdUser, err := userRepo.CreateUserByEmailAndPassword(context.Background(), email, string(passwordHash))
	require.NoError(t, err)

	user, err := userRepo.GetUserByEmail(context.Background(), email)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotNil(t, user.EmailPasswordAccount)
	assert.Equal(t, email, user.EmailPasswordAccount.Email)
	assert.Equal(t, createdUser.ID, user.ID)
}

func TestUserRepository_GetUserByGoogleSub(t *testing.T) {
	cleanupDatabase(t)

	sub := "google-sub-456"
	googleAccount := &models.GoogleAccount{
		Sub:           sub,
		Email:         "google2@example.com",
		EmailVerified: true,
		Name:          "Google User 2",
	}

	createdUser, err := userRepo.CreateUserByGoogle(context.Background(), googleAccount)
	require.NoError(t, err)

	user, err := userRepo.GetUserByGoogleSub(context.Background(), sub)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotNil(t, user.GoogleAccount)
	assert.Equal(t, sub, user.GoogleAccount.Sub)
	assert.Equal(t, createdUser.ID, user.ID)
}

func TestUserRepository_GetUserById(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	createdUser := createTestUser(t, "getbyid@example.com", string(passwordHash))

	user, err := userRepo.GetUserById(context.Background(), createdUser.ID)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, createdUser.ID, user.ID)
}

func TestUserRepository_UserNotFound(t *testing.T) {
	cleanupDatabase(t)

	_, err := userRepo.GetUserByEmail(context.Background(), "notfound@example.com")
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrUserNotFound, err)

	_, err = userRepo.GetUserByGoogleSub(context.Background(), "non-existent-sub")
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrUserNotFound, err)

	_, err = userRepo.GetUserById(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrUserNotFound, err)
}

func TestUserRepository_DuplicateEmail(t *testing.T) {
	cleanupDatabase(t)

	email := "duplicate@example.com"
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	_, err := userRepo.CreateUserByEmailAndPassword(context.Background(), email, string(passwordHash))
	require.NoError(t, err)

	_, err = userRepo.CreateUserByEmailAndPassword(context.Background(), email, string(passwordHash))
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrUserAlreadyExists, err)
}

func TestUserRepository_DuplicateGoogleSub(t *testing.T) {
	cleanupDatabase(t)

	sub := "unique-sub-test"

	googleAccount1 := &models.GoogleAccount{
		Sub:           sub,
		Email:         "user1@example.com",
		EmailVerified: true,
		Name:          "User 1",
	}
	_, err := userRepo.CreateUserByGoogle(context.Background(), googleAccount1)
	require.NoError(t, err)

	googleAccount2 := &models.GoogleAccount{
		Sub:           sub,
		Email:         "user2@example.com",
		EmailVerified: true,
		Name:          "User 2",
	}
	_, err = userRepo.CreateUserByGoogle(context.Background(), googleAccount2)
	assert.Error(t, err)
}

func TestUserRepository_UpdateMetadata(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "metadata@example.com", string(passwordHash))

	metadata := models.JSONB{
		"theme":    "dark",
		"language": "en",
		"settings": map[string]interface{}{
			"notifications": true,
		},
	}

	err := userRepo.UpdateMetadata(context.Background(), user.ID, metadata)
	require.NoError(t, err)

	updatedUser, err := userRepo.GetUserById(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, "dark", updatedUser.Metadata["theme"])
	assert.Equal(t, "en", updatedUser.Metadata["language"])
}

func TestUserRepository_UpdateUserRole(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "role@example.com", string(passwordHash))

	// Initially not admin and not VIP
	assert.False(t, user.IsAdmin)
	assert.False(t, user.IsVip)

	// Update to admin and VIP
	err := userRepo.UpdateUserRole(context.Background(), user.ID, true, true)
	require.NoError(t, err)

	updatedUser, err := userRepo.GetUserById(context.Background(), user.ID)
	require.NoError(t, err)
	assert.True(t, updatedUser.IsAdmin)
	assert.True(t, updatedUser.IsVip)
}

func TestUserRepository_GetAllAdmins(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user1 := createTestUser(t, "admin1@example.com", string(passwordHash))
	user2 := createTestUser(t, "admin2@example.com", string(passwordHash))
	createTestUser(t, "regular@example.com", string(passwordHash))

	// Make two users admins
	userRepo.UpdateUserRole(context.Background(), user1.ID, true, false)
	userRepo.UpdateUserRole(context.Background(), user2.ID, true, false)

	admins, err := userRepo.GetAllAdmins(context.Background())
	require.NoError(t, err)
	assert.Len(t, admins, 2)
}

func TestUserRepository_GetAllVIPs(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user1 := createTestUser(t, "vip1@example.com", string(passwordHash))
	createTestUser(t, "regular@example.com", string(passwordHash))

	// Make one user VIP
	userRepo.UpdateUserRole(context.Background(), user1.ID, false, true)

	vips, err := userRepo.GetAllVIPs(context.Background())
	require.NoError(t, err)
	assert.Len(t, vips, 1)
	assert.Equal(t, user1.ID, vips[0].ID)
}

func TestUserRepository_GetUsersByRole(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user1 := createTestUser(t, "user1@example.com", string(passwordHash))
	user2 := createTestUser(t, "user2@example.com", string(passwordHash))
	user3 := createTestUser(t, "user3@example.com", string(passwordHash))

	userRepo.UpdateUserRole(context.Background(), user1.ID, true, false)
	userRepo.UpdateUserRole(context.Background(), user2.ID, false, true)
	userRepo.UpdateUserRole(context.Background(), user3.ID, true, true)

	// Get admins
	isAdmin := true
	users, total, err := userRepo.GetUsersByRole(context.Background(), &isAdmin, nil, 1, 10)
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, int64(2), total)

	// Get VIPs
	isVIP := true
	users, total, err = userRepo.GetUsersByRole(context.Background(), nil, &isVIP, 1, 10)
	require.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, int64(2), total)
}

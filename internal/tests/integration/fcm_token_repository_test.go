package integration

import (
	"context"
	"testing"
	"time"

	"github.com/nimyab/nim2book-back/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestFcmTokenRepository_AddFcmToken(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "fcm@example.com", string(passwordHash))

	token := "fcm-token-123"
	fcmToken, err := fcmTokenRepo.AddFcmToken(context.Background(), token, user.ID)
	require.NoError(t, err)
	assert.NotNil(t, fcmToken)
	assert.Equal(t, token, fcmToken.Token)
	assert.Equal(t, user.ID, fcmToken.UserID)
	assert.False(t, fcmToken.CreateAt.IsZero())
}

func TestFcmTokenRepository_GetFcmTokensByUserId(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "fcm2@example.com", string(passwordHash))

	// Add multiple tokens
	fcmTokenRepo.AddFcmToken(context.Background(), "token1", user.ID)
	fcmTokenRepo.AddFcmToken(context.Background(), "token2", user.ID)
	fcmTokenRepo.AddFcmToken(context.Background(), "token3", user.ID)

	tokens, err := fcmTokenRepo.GetFcmTokensByUserId(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, tokens, 3)
}

func TestFcmTokenRepository_GetFcmTokensByUserId_Empty(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "fcm3@example.com", string(passwordHash))

	tokens, err := fcmTokenRepo.GetFcmTokensByUserId(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, tokens, 0)
}

func TestFcmTokenRepository_GetFcmTokenByToken(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "fcm4@example.com", string(passwordHash))

	token := "unique-token-456"
	createTestFcmToken(t, token, user)

	fcmToken, err := fcmTokenRepo.GetFcmTokenByToken(context.Background(), token)
	require.NoError(t, err)
	assert.Equal(t, token, fcmToken.Token)
	assert.Equal(t, user.ID, fcmToken.UserID)
}

func TestFcmTokenRepository_GetFcmTokenByToken_NotFound(t *testing.T) {
	cleanupDatabase(t)

	_, err := fcmTokenRepo.GetFcmTokenByToken(context.Background(), "non-existent-token")
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrFcmTokenNotFound, err)
}

func TestFcmTokenRepository_DeleteFcmToken(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "fcm5@example.com", string(passwordHash))

	token := "delete-token-789"
	createTestFcmToken(t, token, user)

	err := fcmTokenRepo.DeleteFcmToken(context.Background(), token, user.ID)
	require.NoError(t, err)

	_, err = fcmTokenRepo.GetFcmTokenByToken(context.Background(), token)
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrFcmTokenNotFound, err)
}

func TestFcmTokenRepository_DeleteFcmToken_NotFound(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "fcm6@example.com", string(passwordHash))

	err := fcmTokenRepo.DeleteFcmToken(context.Background(), "non-existent", user.ID)
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrFcmTokenNotFound, err)
}

func TestFcmTokenRepository_DeleteFcmTokensByUserId(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "fcm7@example.com", string(passwordHash))

	createTestFcmToken(t, "token-a", user)
	createTestFcmToken(t, "token-b", user)
	createTestFcmToken(t, "token-c", user)

	err := fcmTokenRepo.DeleteFcmTokensByUserId(context.Background(), user.ID)
	require.NoError(t, err)

	tokens, err := fcmTokenRepo.GetFcmTokensByUserId(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, tokens, 0)
}

func TestFcmTokenRepository_DuplicateToken(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "fcm8@example.com", string(passwordHash))

	token := "duplicate-token"
	_, err := fcmTokenRepo.AddFcmToken(context.Background(), token, user.ID)
	require.NoError(t, err)

	_, err = fcmTokenRepo.AddFcmToken(context.Background(), token, user.ID)
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrFcmTokenAlreadyAdd, err)
}

func TestFcmTokenRepository_TokenExists(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "fcm9@example.com", string(passwordHash))

	token := "exists-token"
	createTestFcmToken(t, token, user)

	exists, err := fcmTokenRepo.TokenExists(context.Background(), token)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = fcmTokenRepo.TokenExists(context.Background(), "non-existing")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestFcmTokenRepository_CountFcmTokensByUserId(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "fcm10@example.com", string(passwordHash))

	createTestFcmToken(t, "count-token-1", user)
	createTestFcmToken(t, "count-token-2", user)
	createTestFcmToken(t, "count-token-3", user)

	count, err := fcmTokenRepo.CountFcmTokensByUserId(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestFcmTokenRepository_CleanupExpiredTokens(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "fcm11@example.com", string(passwordHash))

	// Add tokens (they will have current timestamp)
	createTestFcmToken(t, "new-token", user)

	// In real scenario, old tokens would have old create_at
	// For testing, we cleanup tokens older than -1 second (which means tokens created before 1 second in the future, i.e., all current tokens)
	deleted, err := fcmTokenRepo.CleanupExpiredTokens(context.Background(), -1*time.Second)
	require.NoError(t, err)
	assert.Equal(t, int64(1), deleted)

	tokens, err := fcmTokenRepo.GetFcmTokensByUserId(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, tokens, 0)
}

func TestFcmTokenRepository_CleanupExpiredTokens_NoExpired(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "fcm12@example.com", string(passwordHash))

	createTestFcmToken(t, "recent-token", user)

	// Cleanup tokens older than 1 year (no tokens should be deleted)
	deleted, err := fcmTokenRepo.CleanupExpiredTokens(context.Background(), 365*24*time.Hour)
	require.NoError(t, err)
	assert.Equal(t, int64(0), deleted)

	tokens, err := fcmTokenRepo.GetFcmTokensByUserId(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, tokens, 1)
}

func TestFcmTokenRepository_CascadeDelete(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "fcm13@example.com", string(passwordHash))

	// Add tokens for user
	createTestFcmToken(t, "cascade-token-1", user)
	createTestFcmToken(t, "cascade-token-2", user)

	// Verify tokens exist
	tokens, err := fcmTokenRepo.GetFcmTokensByUserId(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, tokens, 2)

	// Delete user (should cascade delete tokens)
	err = userRepo.Delete(context.Background(), user.ID)
	require.NoError(t, err)

	// Verify tokens are deleted
	tokens, err = fcmTokenRepo.GetFcmTokensByUserId(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, tokens, 0)
}

func TestFcmTokenRepository_MultipleUsers(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user1 := createTestUser(t, "fcm14@example.com", string(passwordHash))
	user2 := createTestUser(t, "fcm15@example.com", string(passwordHash))

	// Add tokens for both users
	createTestFcmToken(t, "user1-token-1", user1)
	createTestFcmToken(t, "user1-token-2", user1)
	createTestFcmToken(t, "user2-token-1", user2)

	// Verify user1 has 2 tokens
	tokens1, err := fcmTokenRepo.GetFcmTokensByUserId(context.Background(), user1.ID)
	require.NoError(t, err)
	assert.Len(t, tokens1, 2)

	// Verify user2 has 1 token
	tokens2, err := fcmTokenRepo.GetFcmTokensByUserId(context.Background(), user2.ID)
	require.NoError(t, err)
	assert.Len(t, tokens2, 1)

	// Delete user1's tokens
	err = fcmTokenRepo.DeleteFcmTokensByUserId(context.Background(), user1.ID)
	require.NoError(t, err)

	// Verify user1 has no tokens
	tokens1, err = fcmTokenRepo.GetFcmTokensByUserId(context.Background(), user1.ID)
	require.NoError(t, err)
	assert.Len(t, tokens1, 0)

	// Verify user2 still has their token
	tokens2, err = fcmTokenRepo.GetFcmTokensByUserId(context.Background(), user2.ID)
	require.NoError(t, err)
	assert.Len(t, tokens2, 1)
}

func TestFcmTokenRepository_TokenOrdering(t *testing.T) {
	cleanupDatabase(t)

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := createTestUser(t, "fcm16@example.com", string(passwordHash))

	// Add tokens with slight delay
	createTestFcmToken(t, "token-first", user)
	time.Sleep(10 * time.Millisecond)
	createTestFcmToken(t, "token-second", user)
	time.Sleep(10 * time.Millisecond)
	createTestFcmToken(t, "token-third", user)

	// Get tokens (should be ordered by create_at DESC)
	tokens, err := fcmTokenRepo.GetFcmTokensByUserId(context.Background(), user.ID)
	require.NoError(t, err)
	assert.Len(t, tokens, 3)

	// Most recent token should be first
	assert.Equal(t, "token-third", tokens[0].Token)
	assert.Equal(t, "token-second", tokens[1].Token)
	assert.Equal(t, "token-first", tokens[2].Token)
}

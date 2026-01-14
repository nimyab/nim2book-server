package integration

import (
	"context"
	"testing"
	"time"

	"github.com/nimyab/nim2book-back/internal/services/auth/login"
	"github.com/nimyab/nim2book-back/internal/services/auth/register"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthService_Register(t *testing.T) {
	cleanupDatabase(t)

	registerService := register.New(userRepo)

	email := "register@example.com"
	password := "password123"

	output, err := registerService.Register(&register.Input{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.True(t, output.Success)

	// Verify user was created in database
	user, err := userRepo.GetUserByEmail(context.Background(), email)
	require.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotNil(t, user.EmailPasswordAccount)
	assert.Equal(t, email, user.EmailPasswordAccount.Email)
}

func TestAuthService_RegisterDuplicateEmail(t *testing.T) {
	cleanupDatabase(t)

	registerService := register.New(userRepo)

	email := "duplicate@example.com"
	password := "password123"

	// First registration
	output, err := registerService.Register(&register.Input{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.True(t, output.Success)

	// Second registration with same email
	output, err = registerService.Register(&register.Input{
		Email:    email,
		Password: "anotherpassword",
	})
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, register.ErrUserAlreadyExist, err)
}

func TestAuthService_Login(t *testing.T) {
	cleanupDatabase(t)

	registerService := register.New(userRepo)
	loginService := login.New(userRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)

	email := "login@example.com"
	password := "password123"

	// Register user first
	_, err := registerService.Register(&register.Input{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)

	// Login
	output, err := loginService.Login(&login.Input{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotNil(t, output)
	assert.NotEmpty(t, output.AccessToken)
	assert.NotEmpty(t, output.RefreshToken)
	assert.NotNil(t, output.User)
	assert.Equal(t, email, output.User.EmailPasswordAccount.Email)
}

func TestAuthService_LoginWrongPassword(t *testing.T) {
	cleanupDatabase(t)

	registerService := register.New(userRepo)
	loginService := login.New(userRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)

	email := "wrongpass@example.com"
	password := "correctpassword"

	// Register user
	_, err := registerService.Register(&register.Input{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)

	// Login with wrong password
	output, err := loginService.Login(&login.Input{
		Email:    email,
		Password: "wrongpassword",
	})
	assert.Error(t, err)
	assert.Nil(t, output)
	assert.Equal(t, login.ErrPasswordDoNotMatch, err)
}

func TestAuthService_LoginUserNotFound(t *testing.T) {
	cleanupDatabase(t)

	loginService := login.New(userRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)

	output, err := loginService.Login(&login.Input{
		Email:    "notfound@example.com",
		Password: "password123",
	})
	assert.Error(t, err)
	assert.Nil(t, output)
}

func TestAuthService_CompleteAuthFlow(t *testing.T) {
	cleanupDatabase(t)

	registerService := register.New(userRepo)
	loginService := login.New(userRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)

	email := "authflow@example.com"
	password := "password123"

	// Step 1: Register
	registerOutput, err := registerService.Register(&register.Input{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.True(t, registerOutput.Success)

	// Step 2: Login
	loginOutput, err := loginService.Login(&login.Input{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, loginOutput.AccessToken)
	assert.NotEmpty(t, loginOutput.RefreshToken)

	// Step 3: Verify user data
	user := loginOutput.User
	assert.NotNil(t, user)
	assert.False(t, user.IsAdmin)
	assert.False(t, user.IsVip)
	assert.NotNil(t, user.EmailPasswordAccount)
	assert.Equal(t, email, user.EmailPasswordAccount.Email)

	// Verify password is hashed
	assert.NotEqual(t, password, user.EmailPasswordAccount.PasswordHash)
	err = bcrypt.CompareHashAndPassword([]byte(user.EmailPasswordAccount.PasswordHash), []byte(password))
	assert.NoError(t, err)
}

func TestAuthService_PasswordHashing(t *testing.T) {
	cleanupDatabase(t)

	registerService := register.New(userRepo)

	email := "hash@example.com"
	password := "mySecurePassword123"

	_, err := registerService.Register(&register.Input{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)

	// Retrieve user and check password is hashed
	user, err := userRepo.GetUserByEmail(context.Background(), email)
	require.NoError(t, err)

	passwordHash := user.EmailPasswordAccount.PasswordHash
	assert.NotEqual(t, password, passwordHash)
	assert.Greater(t, len(passwordHash), 50) // bcrypt hash is typically 60 chars

	// Verify password can be validated
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
	assert.NoError(t, err)

	// Verify wrong password fails
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("wrongpassword"))
	assert.Error(t, err)
}

func TestAuthService_MultipleUsers(t *testing.T) {
	cleanupDatabase(t)

	registerService := register.New(userRepo)
	loginService := login.New(userRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)

	// Register multiple users
	users := []struct {
		email    string
		password string
	}{
		{"user1@example.com", "password1"},
		{"user2@example.com", "password2"},
		{"user3@example.com", "password3"},
	}

	for _, u := range users {
		output, err := registerService.Register(&register.Input{
			Email:    u.email,
			Password: u.password,
		})
		require.NoError(t, err)
		assert.True(t, output.Success)
	}

	// Login with each user
	for _, u := range users {
		output, err := loginService.Login(&login.Input{
			Email:    u.email,
			Password: u.password,
		})
		require.NoError(t, err)
		assert.Equal(t, u.email, output.User.EmailPasswordAccount.Email)
	}

	// Verify wrong passwords fail
	for _, u := range users {
		output, err := loginService.Login(&login.Input{
			Email:    u.email,
			Password: "wrongpassword",
		})
		assert.Error(t, err)
		assert.Nil(t, output)
	}
}

func TestAuthService_JWTTokenGeneration(t *testing.T) {
	cleanupDatabase(t)

	registerService := register.New(userRepo)
	loginService := login.New(userRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)

	email := "jwt@example.com"
	password := "password123"

	// Register and login
	registerService.Register(&register.Input{
		Email:    email,
		Password: password,
	})

	output, err := loginService.Login(&login.Input{
		Email:    email,
		Password: password,
	})
	require.NoError(t, err)

	// Verify tokens are not empty
	assert.NotEmpty(t, output.AccessToken)
	assert.NotEmpty(t, output.RefreshToken)

	// Verify tokens are different
	assert.NotEqual(t, output.AccessToken, output.RefreshToken)

	// Tokens should be JWT format (3 parts separated by dots)
	assert.Contains(t, output.AccessToken, ".")
	assert.Contains(t, output.RefreshToken, ".")
}

func TestAuthService_EmptyCredentials(t *testing.T) {
	cleanupDatabase(t)

	registerService := register.New(userRepo)
	loginService := login.New(userRepo, "test-secret", 15*time.Minute, 7*24*time.Hour)

	// Register with empty email
	output, err := registerService.Register(&register.Input{
		Email:    "",
		Password: "password",
	})
	assert.Error(t, err)
	assert.Nil(t, output)

	// Login with empty email
	output2, err := loginService.Login(&login.Input{
		Email:    "",
		Password: "password",
	})
	assert.Error(t, err)
	assert.Nil(t, output2)

	// Login with empty password
	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	testEmail := "empty@example.com"
	createTestUser(t, testEmail, string(passwordHash))

	output3, err := loginService.Login(&login.Input{
		Email:    testEmail,
		Password: "",
	})
	assert.Error(t, err)
	assert.Nil(t, output3)
}

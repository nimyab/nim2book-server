package integration

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/joho/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/internal/repositories"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	testDB       *gorm.DB
	testRedis    *redis.Client
	testMinIO    *minio.Client
	userRepo     *repositories.UserRepository
	bookRepo     *repositories.BookRepository
	dictRepo     *repositories.DictionaryRepository
	fcmTokenRepo *repositories.FcmTokenRepository
)

// TestMain sets up and tears down the test environment
func TestMain(m *testing.M) {
	// Load test environment variables
	if err := godotenv.Load("../../../.env.test"); err != nil {
		log.Printf("Warning: .env.test file not found, using environment variables")
	}

	// Wait for services to be ready
	if err := waitForServices(); err != nil {
		log.Fatalf("Failed to wait for services: %v", err)
	}

	// Setup test database
	if err := setupDatabase(); err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Setup Redis
	if err := setupRedis(); err != nil {
		log.Fatalf("Failed to setup redis: %v", err)
	}

	// Setup MinIO
	if err := setupMinIO(); err != nil {
		log.Fatalf("Failed to setup MinIO: %v", err)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	teardown()

	os.Exit(code)
}

func waitForServices() error {
	maxRetries := 30
	retryDelay := time.Second

	// Wait for PostgreSQL
	for i := 0; i < maxRetries; i++ {
		dsn := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_SSLMODE"),
		)
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			sqlDB, _ := db.DB()
			sqlDB.Close()
			log.Println("PostgreSQL is ready")
			break
		}
		if i == maxRetries-1 {
			return fmt.Errorf("PostgreSQL not ready after %d attempts", maxRetries)
		}
		time.Sleep(retryDelay)
	}

	// Wait for Redis
	for i := 0; i < maxRetries; i++ {
		client := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       0,
		})
		if err := client.Ping(context.Background()).Err(); err == nil {
			client.Close()
			log.Println("Redis is ready")
			break
		}
		if i == maxRetries-1 {
			return fmt.Errorf("Redis not ready after %d attempts", maxRetries)
		}
		time.Sleep(retryDelay)
	}

	// Wait for MinIO
	for i := 0; i < maxRetries; i++ {
		minioClient, err := minio.New(os.Getenv("S3_ENDPOINT"), &minio.Options{
			Creds:  credentials.NewStaticV4(os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY"), ""),
			Secure: false,
		})
		if err == nil {
			_, err = minioClient.ListBuckets(context.Background())
			if err == nil {
				log.Println("MinIO is ready")
				break
			}
		}
		if i == maxRetries-1 {
			return fmt.Errorf("MinIO not ready after %d attempts", maxRetries)
		}
		time.Sleep(retryDelay)
	}

	return nil
}

func setupDatabase() error {
	var err error

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE"),
	)

	testDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Enable uuid-ossp extension
	if err := testDB.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`).Error; err != nil {
		return fmt.Errorf("failed to create uuid-ossp extension: %w", err)
	}

	// Run migrations
	err = testDB.AutoMigrate(
		&models.GoogleAccount{},
		&models.EmailPasswordAccount{},
		&models.User{},
		&models.FcmToken{},
		&models.Book{},
		&models.Dictionary{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Initialize repositories
	userRepo = repositories.NewUserRepository(testDB)
	bookRepo = repositories.NewBookRepository(testDB)
	dictRepo = repositories.NewDictionaryRepository(testDB)
	fcmTokenRepo = repositories.NewFcmTokenRepository(testDB)

	log.Println("Database setup completed")
	return nil
}

func setupRedis() error {
	testRedis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	if err := testRedis.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Println("Redis setup completed")
	return nil
}

func setupMinIO() error {
	var err error
	testMinIO, err = minio.New(os.Getenv("S3_ENDPOINT"), &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("S3_ACCESS_KEY"), os.Getenv("S3_SECRET_KEY"), ""),
		Secure: false,
	})
	if err != nil {
		return fmt.Errorf("failed to connect to MinIO: %w", err)
	}

	log.Println("MinIO setup completed")
	return nil
}

func teardown() {
	if testDB != nil {
		sqlDB, _ := testDB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	if testRedis != nil {
		testRedis.Close()
	}

	log.Println("Teardown completed")
}

// cleanupDatabase removes all test data from database
func cleanupDatabase(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	// Delete in correct order due to foreign key constraints
	testDB.Exec("DELETE FROM fcm_tokens")
	testDB.Exec("DELETE FROM users")
	testDB.Exec("DELETE FROM email_password_accounts")
	testDB.Exec("DELETE FROM google_accounts")
	testDB.Exec("DELETE FROM books")
	testDB.Exec("DELETE FROM dictionary")

	// Clear Redis
	if err := testRedis.FlushDB(ctx).Err(); err != nil {
		t.Logf("Warning: failed to flush redis: %v", err)
	}
}

// createTestUser creates a test user with email/password
func createTestUser(t *testing.T, email, passwordHash string) *models.User {
	t.Helper()

	user, err := userRepo.CreateUserByEmailAndPassword(context.Background(), email, passwordHash)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// createTestUserGoogle creates a test user with Google account
func createTestUserGoogle(t *testing.T, sub, email, name string) *models.User {
	t.Helper()

	googleAccount := &models.GoogleAccount{
		Sub:           sub,
		Email:         email,
		EmailVerified: true,
		Name:          name,
		Picture:       nil,
	}

	user, err := userRepo.CreateUserByGoogle(context.Background(), googleAccount)
	if err != nil {
		t.Fatalf("Failed to create test user with Google: %v", err)
	}

	return user
}

// createTestBook creates a test book
func createTestBook(t *testing.T, title, author string, chapterPaths []string) *models.Book {
	t.Helper()

	book := &models.Book{
		Title:        title,
		Author:       author,
		ChapterPaths: models.StringArray(chapterPaths),
		Cover:        nil,
	}

	createdBook, err := bookRepo.CreateBook(context.Background(), book)
	if err != nil {
		t.Fatalf("Failed to create test book: %v", err)
	}

	return createdBook
}

// createTestDictionary creates a test dictionary entry
func createTestDictionary(t *testing.T, text, lang string, content []byte) *models.Dictionary {
	t.Helper()

	dict, err := dictRepo.CreateDictionaryData(context.Background(), text, lang, content)
	if err != nil {
		t.Fatalf("Failed to create test dictionary: %v", err)
	}

	return dict
}

// createTestFcmToken creates a test FCM token
func createTestFcmToken(t *testing.T, token string, user *models.User) *models.FcmToken {
	t.Helper()

	fcmToken, err := fcmTokenRepo.AddFcmToken(context.Background(), token, user.ID)
	if err != nil {
		t.Fatalf("Failed to create test FCM token: %v", err)
	}

	return fcmToken
}

package postgres_gorm

import (
	"context"
	"fmt"
	"time"

	"github.com/nimyab/nim2book-back/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	PostgresURL string
}

type Postgres struct {
	DB *gorm.DB
}

func New(ctx context.Context, cfg *Config) (*Postgres, error) {
	const operation = "postgres_gorm.New"

	// Configure GORM
	gormConfig := &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Info),
		PrepareStmt:            true,
		SkipDefaultTransaction: false,
		CreateBatchSize:        100,
	}

	db, err := gorm.Open(postgres.Open(cfg.PostgresURL), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	// Ping database to verify connection
	if err = sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("%s: unable to ping database: %w", operation, err)
	}

	// Initialize adapter with repositories
	p := &Postgres{
		DB: db,
	}

	return p, nil
}

func (p *Postgres) Close() error {
	if p.DB != nil {
		sqlDB, err := p.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

// AutoMigrate runs auto migration for all models
func (p *Postgres) AutoMigrate() error {
	return p.DB.AutoMigrate(
		&models.GoogleAccount{},
		&models.EmailPasswordAccount{},
		&models.User{},
		&models.FcmToken{},
		&models.Book{},
		&models.Dictionary{},
	)
}

// HealthCheck checks database connection health
func (p *Postgres) HealthCheck(ctx context.Context) error {
	sqlDB, err := p.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

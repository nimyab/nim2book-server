package app

import (
	"context"

	"firebase.google.com/go/v4/messaging"
	"github.com/nimyab/nim2book-back/config"
	"github.com/nimyab/nim2book-back/internal/adapter/firebase"
	"github.com/nimyab/nim2book-back/internal/adapter/minio"
	"github.com/nimyab/nim2book-back/internal/adapter/postgres_sqlc"
	"github.com/nimyab/nim2book-back/internal/adapter/redis_cache"
	"github.com/nimyab/nim2book-back/internal/services/word_aligner"
	pb "github.com/nimyab/nim2book-back/proto/word_aligner"
	"github.com/samber/do/v2"
)

// registerAdapters registers all infrastructure adapters (database, storage, cache, etc.)
func (a *App) registerAdapters() error {
	// PostgreSQL
	do.Provide(a.injector, func(i do.Injector) (*postgres_sqlc.Postgres, error) {
		cfg := do.MustInvoke[*config.Config](i)
		return postgres_sqlc.New(context.Background(), &postgres_sqlc.Config{
			PostgresURL: cfg.PostgresURL,
		})
	})

	// MinIO
	do.Provide(a.injector, func(i do.Injector) (*minio.Minio, error) {
		cfg := do.MustInvoke[*config.Config](i)
		return minio.New(context.Background(), &minio.Config{
			MinioURL:          cfg.MinioURL,
			MinioRootUser:     cfg.MinioRootUser,
			MinioRootPassword: cfg.MinioRootPassword,
			MinioBucketName:   cfg.MinioBucketName,
			MinioRegion:       cfg.MinioRegion,
			MinioUseSSL:       cfg.MinioUseSSL,
		})
	})

	// Redis Cache
	do.Provide(a.injector, func(i do.Injector) (*redis_cache.RedisCache, error) {
		cfg := do.MustInvoke[*config.Config](i)
		return redis_cache.New(&redis_cache.Config{
			RedisURL: cfg.RedisURL,
		})
	})

	// Firebase Messaging
	do.Provide(a.injector, func(i do.Injector) (*messaging.Client, error) {
		cfg := do.MustInvoke[*config.Config](i)
		firebaseApp, err := firebase.New(context.Background(), &firebase.Config{
			GoogleCredentials: cfg.GoogleCredentials,
		})
		if err != nil {
			return nil, err
		}
		return firebaseApp.Messaging(context.Background())
	})

	// Word Aligner gRPC Client
	do.Provide(a.injector, func(i do.Injector) (pb.AlignmentServiceClient, error) {
		cfg := do.MustInvoke[*config.Config](i)
		return word_aligner.NewClient(&word_aligner.ClientConfig{
			Address: cfg.WordAlignerAddrGrpc,
		})
	})

	return nil
}

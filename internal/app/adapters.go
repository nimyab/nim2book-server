package app

import (
	"context"

	"firebase.google.com/go/v4/messaging"
	"github.com/nimyab/nim2book-back/config"
	"github.com/nimyab/nim2book-back/ent"
	"github.com/nimyab/nim2book-back/internal/adapter/firebase"
	"github.com/nimyab/nim2book-back/internal/adapter/minio"
	"github.com/nimyab/nim2book-back/internal/adapter/postgres"
	"github.com/nimyab/nim2book-back/internal/adapter/redis_cache"
	"github.com/nimyab/nim2book-back/internal/services/word_aligner"
)

type Adapters struct {
	EntClient   *ent.Client
	Minio       *minio.Minio
	Redis       *redis_cache.RedisCache
	Firebase    *messaging.Client
	WordAligner *word_aligner.Client
}

func newAdapters(cfg *config.Config) (*Adapters, error) {
	// PostgreSQL
	entClient, err := postgres.New(&postgres.Config{
		PostgresURL: cfg.PostgresURL,
		IsDebug:     cfg.Env == config.EnvLocal || cfg.Env == config.EnvDev,
	})
	if err != nil {
		return nil, err
	}

	// MinIO
	minioClient, err := minio.New(context.Background(), &minio.Config{
		MinioURL:          cfg.MinioURL,
		MinioRootUser:     cfg.MinioRootUser,
		MinioRootPassword: cfg.MinioRootPassword,
		MinioBucketName:   cfg.MinioBucketName,
		MinioRegion:       cfg.MinioRegion,
		MinioUseSSL:       cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, err
	}

	// Redis Cache
	redisClient, err := redis_cache.New(&redis_cache.Config{
		RedisURL: cfg.RedisURL,
	})
	if err != nil {
		return nil, err
	}

	// Firebase Messaging
	firebaseApp, err := firebase.New(context.Background(), &firebase.Config{
		GoogleCredentials: cfg.GoogleCredentials,
	})
	if err != nil {
		return nil, err
	}
	firebaseMessaging, err := firebaseApp.Messaging(context.Background())
	if err != nil {
		return nil, err
	}

	// Word Aligner gRPC Client
	wordAlignerClient, err := word_aligner.NewClient(&word_aligner.ClientConfig{
		Address: cfg.WordAlignerAddrGrpc,
	})
	if err != nil {
		return nil, err
	}

	return &Adapters{
		EntClient:   entClient,
		Minio:       minioClient,
		Redis:       redisClient,
		Firebase:    firebaseMessaging,
		WordAligner: wordAlignerClient,
	}, nil
}

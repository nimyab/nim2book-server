package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nimyab/nim2book-back/config"
	"github.com/nimyab/nim2book-back/internal/adapter/postgres"
	"github.com/nimyab/nim2book-back/internal/adapter/redis"
	"github.com/nimyab/nim2book-back/internal/adapter/s3"
	"github.com/nimyab/nim2book-back/internal/controller/http"
	"github.com/nimyab/nim2book-back/internal/libretranslate/translate"
	"github.com/nimyab/nim2book-back/internal/translate/translate_book"
	"github.com/nimyab/nim2book-back/internal/word_aligner/align"
	"github.com/nimyab/nim2book-back/pkg/logger"
)

// @title						Nim2Book api
// @version					1.0
// @BasePath					/api/v1
// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {
	cfg := config.GetConfig()
	fmt.Println(cfg)

	slogLogger := logger.New(logger.Config{
		Env: cfg.Env,
	})
	slog.SetDefault(slogLogger)

	err := appRun(cfg)
	if err != nil {
		panic(err)
	}
}

func appRun(cfg *config.Config) error {
	// Postgres
	pgClient, err := postgres.New(context.Background(), &postgres.Config{
		PostgresURL: cfg.PostgresURL,
	})
	if err != nil {
		return err
	}

	// S3 storage
	s3Client, err := s3.New(&s3.Config{
		S3URL:          cfg.S3URL,
		S3RootUser:     cfg.S3RootUser,
		S3RootPassword: cfg.S3RootPassword,
		S3BucketName:   cfg.S3BucketName,
		S3Region:       cfg.S3Region,
	})
	if err != nil {
		return err
	}

	// Redis
	redisClient, err := redis.New(&redis.Config{
		RedisURL: cfg.RedisURL,
	})
	if err != nil {
		return err
	}
	_ = redisClient // todo

	// services
	wordAlign := align.New(cfg.WordAlignerURL)
	translateService := translate.New(cfg.LibreTranslateURL)
	translate_book.New(s3Client, pgClient, wordAlign, translateService)

	router := http.Router()
	if err = router.Start(cfg.Port); err != nil {
		return err
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig

	pgClient.Close()
	return nil
}

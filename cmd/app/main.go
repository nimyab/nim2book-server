package main

import (
	"context"
	"github.com/nimyab/nim2book-back/internal/auth/login"
	"github.com/nimyab/nim2book-back/internal/auth/refresh"
	"github.com/nimyab/nim2book-back/internal/auth/register"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nimyab/nim2book-back/config"
	"github.com/nimyab/nim2book-back/internal/adapter/postgres"
	"github.com/nimyab/nim2book-back/internal/adapter/redis"
	"github.com/nimyab/nim2book-back/internal/adapter/s3"
	"github.com/nimyab/nim2book-back/internal/book/get_book"
	"github.com/nimyab/nim2book-back/internal/book/get_books"
	"github.com/nimyab/nim2book-back/internal/book/get_chapter"
	"github.com/nimyab/nim2book-back/internal/controller/http"
	"github.com/nimyab/nim2book-back/internal/libretranslate/translate"
	"github.com/nimyab/nim2book-back/internal/translate/translate_book"
	"github.com/nimyab/nim2book-back/internal/word_aligner/align"
	"github.com/nimyab/nim2book-back/pkg/logger"
)

// @title						Nim2Book api
// @version						1.0
// @BasePath					/api/v1
// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func main() {
	cfg := config.GetConfig()

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

	// align service
	wordAlign := align.New(cfg.WordAlignerURL)

	// libretranslate service
	translateService := translate.New(cfg.LibreTranslateURL)

	// translate service
	translate_book.New(s3Client, pgClient, wordAlign, translateService, cfg.MaxRequestCount)

	// book service
	get_chapter.New(s3Client)
	get_books.New(pgClient)
	get_book.New(pgClient)

	// auth service
	register.New(pgClient)
	login.New(pgClient, cfg.JWTSecret, cfg.JWTAccessTime, cfg.JWTRefreshTime)
	refresh.New(cfg.JWTSecret, cfg.JWTAccessTime, cfg.JWTRefreshTime)

	router := http.Router()
	if err = router.Start(cfg.Port); err != nil {
		return err
	}

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig

	pgClient.Close()
	return nil
}

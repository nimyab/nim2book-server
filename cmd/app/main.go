package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/nimyab/nim2book-back/internal/adapter/firebase"
	"github.com/nimyab/nim2book-back/internal/adapter/rabbitmq"
	"github.com/nimyab/nim2book-back/internal/auth/google_login"
	"github.com/nimyab/nim2book-back/internal/auth/login"
	"github.com/nimyab/nim2book-back/internal/auth/refresh"
	"github.com/nimyab/nim2book-back/internal/auth/register"
	"github.com/nimyab/nim2book-back/internal/book/update_book"
	"github.com/nimyab/nim2book-back/internal/controller/websocket"
	"github.com/nimyab/nim2book-back/internal/dictionary/lookup"
	"github.com/nimyab/nim2book-back/internal/fcm_token/add_fcm_token"
	"github.com/nimyab/nim2book-back/internal/fcm_token/delete_fcm_token"
	"github.com/nimyab/nim2book-back/internal/file/file_public"
	"github.com/nimyab/nim2book-back/internal/user/me"

	"github.com/nimyab/nim2book-back/config"
	"github.com/nimyab/nim2book-back/internal/adapter/postgres"
	"github.com/nimyab/nim2book-back/internal/adapter/redis_cache"
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
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
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

	// RedisCache
	redisCacheClient, err := redis_cache.New(&redis_cache.Config{
		RedisURL: cfg.RedisURL,
	})
	if err != nil {
		return err
	}

	// RabbitMQ
	rabbit, err := rabbitmq.New(&rabbitmq.Config{
		RabbitmqUrl: cfg.RabbitmqUrl,
	})
	if err != nil {
		return err
	}

	// Firebase
	firebaseApp, err := firebase.New(context.Background(), &firebase.Config{GoogleCredentials: cfg.GoogleCredentials})
	if err != nil {
		return err
	}
	messagingFirebaseClient, err := firebaseApp.Messaging(context.Background())
	if err != nil {
		return err
	}

	// align service
	wordAlign := align.New(cfg.WordAlignerURL)

	// libretranslate service
	translateService := translate.New(cfg.LibreTranslateURL)

	// translate service
	translate_book.New(
		s3Client,
		pgClient,
		wordAlign,
		translateService,
		rabbit,
		cfg.MaxRequestCount,
		cfg.WaitMilliseconds,
		messagingFirebaseClient,
	)

	// book service
	get_chapter.New(s3Client)
	get_books.New(pgClient)
	get_book.New(pgClient)
	update_book.New(pgClient, s3Client)

	// auth service
	register.New(pgClient)
	login.New(pgClient, cfg.JWTSecret, cfg.JWTAccessTime, cfg.JWTRefreshTime)
	google_login.New(pgClient, cfg.GoogleClientId, cfg.JWTSecret, cfg.JWTAccessTime, cfg.JWTRefreshTime)
	refresh.New(cfg.JWTSecret, cfg.JWTAccessTime, cfg.JWTRefreshTime)

	// user service
	me.New(pgClient)

	// dictionary service
	lookup.New(pgClient, redisCacheClient, cfg.YandexDictionaryKey, cfg.YandexDictionaryURL)

	// file services
	file_public.New(s3Client)

	// fcm_token services
	add_fcm_token.New(pgClient)
	delete_fcm_token.New(pgClient)

	// websocket
	websocket.NewAndStart()

	router := http.Router(cfg.JWTSecret)

	go func() {
		if err = router.Start(cfg.Port); err != nil {
			slog.Error(err.Error())
			return
		}
	}()

	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	<-sig

	_ = router.Shutdown(context.Background())
	pgClient.Close()
	if err = rabbit.Close(); err != nil {
		slog.Error(err.Error())
	}

	return nil
}

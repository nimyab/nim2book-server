package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/maniartech/signals"
	"github.com/nimyab/nim2book-back/config"
	"github.com/nimyab/nim2book-back/internal/adapter/firebase"
	"github.com/nimyab/nim2book-back/internal/adapter/minio"
	"github.com/nimyab/nim2book-back/internal/adapter/postgres_gorm"
	"github.com/nimyab/nim2book-back/internal/adapter/redis_cache"
	"github.com/nimyab/nim2book-back/internal/controller/http"
	"github.com/nimyab/nim2book-back/internal/controller/websocket"
	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/internal/repositories"
	"github.com/nimyab/nim2book-back/internal/services/auth/google_login"
	"github.com/nimyab/nim2book-back/internal/services/auth/login"
	"github.com/nimyab/nim2book-back/internal/services/auth/refresh"
	"github.com/nimyab/nim2book-back/internal/services/auth/register"
	"github.com/nimyab/nim2book-back/internal/services/book/get_book"
	"github.com/nimyab/nim2book-back/internal/services/book/get_books"
	"github.com/nimyab/nim2book-back/internal/services/book/get_chapter"
	"github.com/nimyab/nim2book-back/internal/services/book/update_book"
	"github.com/nimyab/nim2book-back/internal/services/dictionary/lookup"
	"github.com/nimyab/nim2book-back/internal/services/fcm_token/add_fcm_token"
	"github.com/nimyab/nim2book-back/internal/services/fcm_token/delete_fcm_token"
	"github.com/nimyab/nim2book-back/internal/services/file/file_public"
	"github.com/nimyab/nim2book-back/internal/services/libretranslate/translate"
	"github.com/nimyab/nim2book-back/internal/services/notification"
	"github.com/nimyab/nim2book-back/internal/services/translate/translate_book"
	"github.com/nimyab/nim2book-back/internal/services/user/me"
	"github.com/nimyab/nim2book-back/internal/services/user/metadata"
	"github.com/nimyab/nim2book-back/internal/services/word_aligner"
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
	// register signals for messaging
	notificationSignal := signals.New[models.Notification]()

	// Postgres
	pgClient, err := postgres_gorm.New(
		context.Background(),
		&postgres_gorm.Config{PostgresURL: cfg.PostgresURL},
	)
	if err != nil {
		return err
	}

	// MinIO storage
	minioClient, err := minio.New(context.Background(), &minio.Config{
		MinioURL:          cfg.MinioURL,
		MinioRootUser:     cfg.MinioRootUser,
		MinioRootPassword: cfg.MinioRootPassword,
		MinioBucketName:   cfg.MinioBucketName,
		MinioRegion:       cfg.MinioRegion,
		MinioUseSSL:       cfg.MinioUseSSL,
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

	// Firebase
	firebaseApp, err := firebase.New(context.Background(), &firebase.Config{GoogleCredentials: cfg.GoogleCredentials})
	if err != nil {
		return err
	}
	messagingFirebaseClient, err := firebaseApp.Messaging(context.Background())
	if err != nil {
		return err
	}

	// repositories
	userRepository := repositories.NewUserRepository(pgClient.DB)
	bookRepository := repositories.NewBookRepository(pgClient.DB)
	fcmTokensRepository := repositories.NewFcmTokenRepository(pgClient.DB)
	dictionaryRepository := repositories.NewDictionaryRepository(pgClient.DB)

	// notification service
	notificationService := notification.New(messagingFirebaseClient, fcmTokensRepository)

	// align service
	wordAlignerClient, err := word_aligner.NewClient(&word_aligner.ClientConfig{Address: cfg.WordAlignerAddrGrpc})
	if err != nil {
		return err
	}

	// libretranslate service
	translateService := translate.New(cfg.LibreTranslateURL)

	// translate service
	translate_book.New(
		minioClient,
		bookRepository,
		wordAlignerClient,
		translateService,
		cfg.MaxRequestCount,
		cfg.WaitMilliseconds,
		notificationSignal,
	)

	// book service
	get_chapter.New(minioClient)
	get_books.New(bookRepository)
	get_book.New(bookRepository)
	update_book.New(bookRepository, minioClient)

	// auth service
	register.New(userRepository)
	login.New(userRepository, cfg.JWTSecret, cfg.JWTAccessTime, cfg.JWTRefreshTime)
	google_login.New(userRepository, cfg.GoogleClientId, cfg.JWTSecret, cfg.JWTAccessTime, cfg.JWTRefreshTime)
	refresh.New(cfg.JWTSecret, cfg.JWTAccessTime, cfg.JWTRefreshTime)

	// user service
	me.New(userRepository)
	metadata.New(userRepository)

	// dictionary service
	lookup.New(dictionaryRepository, redisCacheClient, cfg.YandexDictionaryKey, cfg.YandexDictionaryURL)

	// file services
	file_public.New(minioClient)

	// fcm_token services
	add_fcm_token.New(fcmTokensRepository)
	delete_fcm_token.New(fcmTokensRepository)

	// websocket
	websocket.NewAndStart()

	// signals
	notificationSignal.AddListener(notificationService.ProcessNotification)

	// http
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
	_ = pgClient.Close()

	slog.Info("Graceful shutdown!")

	return nil
}

package app

import (
	"firebase.google.com/go/v4/messaging"
	"github.com/nimyab/nim2book-back/config"
	"github.com/nimyab/nim2book-back/internal/adapter/minio"
	"github.com/nimyab/nim2book-back/internal/adapter/redis_cache"
	"github.com/nimyab/nim2book-back/internal/repository"
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
	"github.com/nimyab/nim2book-back/internal/services/genre/create_genre"
	"github.com/nimyab/nim2book-back/internal/services/genre/delete_genre"
	"github.com/nimyab/nim2book-back/internal/services/genre/get_genre"
	"github.com/nimyab/nim2book-back/internal/services/genre/get_genres"
	"github.com/nimyab/nim2book-back/internal/services/genre/update_genre"
	"github.com/nimyab/nim2book-back/internal/services/libretranslate/translate"
	"github.com/nimyab/nim2book-back/internal/services/notification"
	"github.com/nimyab/nim2book-back/internal/services/personal_user_book/get_personal_user_book"
	"github.com/nimyab/nim2book-back/internal/services/personal_user_book/get_personal_user_books"
	"github.com/nimyab/nim2book-back/internal/services/personal_user_book/update_personal_user_book"
	"github.com/nimyab/nim2book-back/internal/services/translate/translate_book"
	"github.com/nimyab/nim2book-back/internal/services/translate/translate_personal_user_book"
	"github.com/nimyab/nim2book-back/internal/services/user/me"
	"github.com/nimyab/nim2book-back/internal/services/user/metadata"
	pb "github.com/nimyab/nim2book-back/proto/word_aligner"
	"github.com/samber/do/v2"
)

// registerServices registers all business logic services (use cases)
func (a *App) registerServices() {
	// Book services
	do.Provide(a.injector, func(i do.Injector) (*get_books.Service, error) {
		bookRepo := do.MustInvoke[*repository.BookRepository](i)
		return get_books.New(bookRepo), nil
	})

	do.Provide(a.injector, func(i do.Injector) (*get_book.Service, error) {
		bookRepo := do.MustInvoke[*repository.BookRepository](i)
		return get_book.New(bookRepo), nil
	})

	do.Provide(a.injector, func(i do.Injector) (*get_chapter.Service, error) {
		s3 := do.MustInvoke[*minio.Minio](i)
		return get_chapter.New(s3), nil
	})

	do.Provide(a.injector, func(i do.Injector) (*update_book.Service, error) {
		bookRepo := do.MustInvoke[*repository.BookRepository](i)
		authorRepo := do.MustInvoke[*repository.AuthorRepository](i)
		s3 := do.MustInvoke[*minio.Minio](i)
		return update_book.New(bookRepo, authorRepo, s3), nil
	})

	// Genre services
	do.Provide(a.injector, func(i do.Injector) (*get_genres.Service, error) {
		genreRepo := do.MustInvoke[*repository.GenreRepository](i)
		return get_genres.New(genreRepo), nil
	})

	do.Provide(a.injector, func(i do.Injector) (*get_genre.Service, error) {
		genreRepo := do.MustInvoke[*repository.GenreRepository](i)
		return get_genre.New(genreRepo), nil
	})

	do.Provide(a.injector, func(i do.Injector) (*create_genre.Service, error) {
		genreRepo := do.MustInvoke[*repository.GenreRepository](i)
		return create_genre.New(genreRepo), nil
	})

	do.Provide(a.injector, func(i do.Injector) (*update_genre.Service, error) {
		genreRepo := do.MustInvoke[*repository.GenreRepository](i)
		return update_genre.New(genreRepo), nil
	})

	do.Provide(a.injector, func(i do.Injector) (*delete_genre.Service, error) {
		genreRepo := do.MustInvoke[*repository.GenreRepository](i)
		return delete_genre.New(genreRepo), nil
	})

	// Personal User Book services
	do.Provide(a.injector, func(i do.Injector) (*get_personal_user_books.Service, error) {
		personalBookRepo := do.MustInvoke[*repository.PersonalBookRepository](i)
		return get_personal_user_books.New(personalBookRepo), nil
	})

	do.Provide(a.injector, func(i do.Injector) (*get_personal_user_book.Service, error) {
		personalBookRepo := do.MustInvoke[*repository.PersonalBookRepository](i)
		return get_personal_user_book.New(personalBookRepo), nil
	})

	do.Provide(a.injector, func(i do.Injector) (*update_personal_user_book.Service, error) {
		personalBookRepo := do.MustInvoke[*repository.PersonalBookRepository](i)
		authorRepo := do.MustInvoke[*repository.AuthorRepository](i)
		s3 := do.MustInvoke[*minio.Minio](i)
		return update_personal_user_book.New(personalBookRepo, authorRepo, s3), nil
	})

	// Auth services
	do.Provide(a.injector, func(i do.Injector) (*register.Service, error) {
		userRepo := do.MustInvoke[*repository.UserRepository](i)
		return register.New(userRepo), nil
	})

	do.Provide(a.injector, func(i do.Injector) (*login.Service, error) {
		userRepo := do.MustInvoke[*repository.UserRepository](i)
		cfg := do.MustInvoke[*config.Config](i)
		return login.New(userRepo, cfg.JWTSecret, cfg.JWTAccessTime, cfg.JWTRefreshTime), nil
	})

	do.Provide(a.injector, func(i do.Injector) (*google_login.Service, error) {
		userRepo := do.MustInvoke[*repository.UserRepository](i)
		cfg := do.MustInvoke[*config.Config](i)
		return google_login.New(userRepo, cfg.GoogleClientId, cfg.JWTSecret, cfg.JWTAccessTime, cfg.JWTRefreshTime), nil
	})

	do.Provide(a.injector, func(i do.Injector) (*refresh.Service, error) {
		cfg := do.MustInvoke[*config.Config](i)
		return refresh.New(cfg.JWTSecret, cfg.JWTAccessTime, cfg.JWTRefreshTime), nil
	})

	// User services
	do.Provide(a.injector, func(i do.Injector) (*me.Service, error) {
		userRepo := do.MustInvoke[*repository.UserRepository](i)
		return me.New(userRepo), nil
	})

	do.Provide(a.injector, func(i do.Injector) (*metadata.Service, error) {
		userRepo := do.MustInvoke[*repository.UserRepository](i)
		return metadata.New(userRepo), nil
	})

	// Dictionary service
	do.Provide(a.injector, func(i do.Injector) (*lookup.Service, error) {
		dictianaryRepo := do.MustInvoke[*repository.DictionaryRepository](i)
		redis := do.MustInvoke[*redis_cache.RedisCache](i)
		cfg := do.MustInvoke[*config.Config](i)
		return lookup.New(dictianaryRepo, redis, cfg.YandexDictionaryKey, cfg.YandexDictionaryURL), nil
	})

	// FCM Token services
	do.Provide(a.injector, func(i do.Injector) (*add_fcm_token.Service, error) {
		fcmTokenRepo := do.MustInvoke[*repository.FcmTokenRepository](i)
		return add_fcm_token.New(fcmTokenRepo), nil
	})

	do.Provide(a.injector, func(i do.Injector) (*delete_fcm_token.Service, error) {
		fcmTokenRepo := do.MustInvoke[*repository.FcmTokenRepository](i)
		return delete_fcm_token.New(fcmTokenRepo), nil
	})

	// File service
	do.Provide(a.injector, func(i do.Injector) (*file_public.Service, error) {
		s3 := do.MustInvoke[*minio.Minio](i)
		return file_public.New(s3), nil
	})

	// Notification service
	do.Provide(a.injector, func(i do.Injector) (*notification.Service, error) {
		messagingClient := do.MustInvoke[*messaging.Client](i)
		fcmTokenRepo := do.MustInvoke[*repository.FcmTokenRepository](i)
		return notification.New(messagingClient, fcmTokenRepo), nil
	})

	// LibreTranslate service
	do.Provide(a.injector, func(i do.Injector) (*translate.Service, error) {
		cfg := do.MustInvoke[*config.Config](i)
		return translate.New(cfg.LibreTranslateURL), nil
	})

	// Translate Book service
	do.Provide(a.injector, func(i do.Injector) (*translate_book.Service, error) {
		s3 := do.MustInvoke[*minio.Minio](i)
		bookRepo := do.MustInvoke[*repository.BookRepository](i)
		authorRepo := do.MustInvoke[*repository.AuthorRepository](i)
		wordAligner := do.MustInvoke[pb.AlignmentServiceClient](i)
		translator := do.MustInvoke[*translate.Service](i)
		notificationSvc := do.MustInvoke[*notification.Service](i)
		cfg := do.MustInvoke[*config.Config](i)

		return translate_book.New(
			s3,
			bookRepo,
			authorRepo,
			wordAligner,
			translator,
			cfg.MaxRequestCount,
			cfg.WaitMilliseconds,
			notificationSvc,
		), nil
	})

	// Translate Personal User Book service
	do.Provide(a.injector, func(i do.Injector) (*translate_personal_user_book.Service, error) {
		s3 := do.MustInvoke[*minio.Minio](i)
		personalBookRepo := do.MustInvoke[*repository.PersonalBookRepository](i)
		authorRepo := do.MustInvoke[*repository.AuthorRepository](i)
		wordAligner := do.MustInvoke[pb.AlignmentServiceClient](i)
		translator := do.MustInvoke[*translate.Service](i)
		notificationSvc := do.MustInvoke[*notification.Service](i)
		cfg := do.MustInvoke[*config.Config](i)

		return translate_personal_user_book.New(
			s3,
			personalBookRepo,
			authorRepo,
			wordAligner,
			translator,
			cfg.MaxRequestCount,
			cfg.WaitMilliseconds,
			notificationSvc,
		), nil
	})
}

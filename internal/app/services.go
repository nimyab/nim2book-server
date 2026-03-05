package app

import (
	"net/http"

	"github.com/nimyab/nim2book-back/config"
	"github.com/nimyab/nim2book-back/internal/controller/websocket"
	"github.com/nimyab/nim2book-back/internal/domain"
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
	"github.com/nimyab/nim2book-back/internal/services/libretranslate/translate"
	"github.com/nimyab/nim2book-back/internal/services/notification"
	"github.com/nimyab/nim2book-back/internal/services/personal_user_book/get_personal_user_book"
	"github.com/nimyab/nim2book-back/internal/services/personal_user_book/get_personal_user_books"
	"github.com/nimyab/nim2book-back/internal/services/personal_user_book/update_personal_user_book"
	"github.com/nimyab/nim2book-back/internal/services/translate/dto"
	"github.com/nimyab/nim2book-back/internal/services/translate/translate_book"
	"github.com/nimyab/nim2book-back/internal/services/translate/translate_personal_book"
	"github.com/nimyab/nim2book-back/internal/services/user/me"
	"github.com/nimyab/nim2book-back/internal/services/user/metadata"
)

type Services struct {
	GetBooks               *get_books.Service
	GetBook                *get_book.Service
	GetChapter             *get_chapter.Service
	UpdateBook             *update_book.Service
	GetGenres              *get_genres.Service
	GetGenre               *get_genre.Service
	CreateGenre            *create_genre.Service
	DeleteGenre            *delete_genre.Service
	GetPersonalUserBooks   *get_personal_user_books.Service
	GetPersonalUserBook    *get_personal_user_book.Service
	UpdatePersonalUserBook *update_personal_user_book.Service
	Register               *register.Service
	Login                  *login.Service
	GoogleLogin            *google_login.Service
	Refresh                *refresh.Service
	Me                     *me.Service
	Metadata               *metadata.Service
	Lookup                 *lookup.Service
	AddFcmToken            *add_fcm_token.Service
	DeleteFcmToken         *delete_fcm_token.Service
	FilePublic             *file_public.Service
	Notification           *notification.Service
	Translate              *translate.Service
	TranslateBook          *translate_book.Service
	TranslatePersonalBook  *translate_personal_book.Service
}

func newServices(repos *Repositories, adapters *Adapters, cfg *config.Config) *Services {
	// Notification service
	notificationService := notification.New(adapters.Firebase, repos.FcmToken, &wsSender{})

	// LibreTranslate service
	translator := translate.New(cfg.LibreTranslateURL, nil)

	return &Services{
		GetBooks:               get_books.New(repos.Book),
		GetBook:                get_book.New(repos.Book),
		GetChapter:             get_chapter.New(adapters.Minio),
		UpdateBook:             update_book.New(repos.Book, repos.Author, adapters.Minio),
		GetGenres:              get_genres.New(repos.Genre),
		GetGenre:               get_genre.New(repos.Genre),
		CreateGenre:            create_genre.New(repos.Genre),
		DeleteGenre:            delete_genre.New(repos.Genre),
		GetPersonalUserBooks:   get_personal_user_books.New(repos.PersonalBook),
		GetPersonalUserBook:    get_personal_user_book.New(repos.PersonalBook),
		UpdatePersonalUserBook: update_personal_user_book.New(repos.PersonalBook, repos.Author, adapters.Minio),
		Register:               register.New(repos.User),
		Login:                  login.New(repos.User, cfg.JWTSecret, cfg.JWTAccessTime, cfg.JWTRefreshTime),
		GoogleLogin:            google_login.New(repos.User, cfg.GoogleClientId, cfg.JWTSecret, cfg.JWTAccessTime, cfg.JWTRefreshTime),
		Refresh:                refresh.New(cfg.JWTSecret, cfg.JWTAccessTime, cfg.JWTRefreshTime),
		Me:                     me.New(repos.User),
		Metadata:               metadata.New(repos.User),
		Lookup:                 lookup.New(repos.Dictionary, adapters.Redis, http.DefaultClient, cfg.YandexDictionaryKey, cfg.YandexDictionaryURL),
		AddFcmToken:            add_fcm_token.New(repos.FcmToken),
		DeleteFcmToken:         delete_fcm_token.New(repos.FcmToken),
		FilePublic:             file_public.New(adapters.Minio),
		Notification:           notificationService,
		Translate:              translator,
		TranslateBook:          translate_book.New(adapters.Minio, repos.Book, repos.Author, adapters.WordAligner, translator, dto.Config{WaitDuration: cfg.WaitMilliseconds, MaxRequestCount: cfg.MaxRequestCount}, notificationService),
		TranslatePersonalBook:  translate_personal_book.New(adapters.Minio, repos.PersonalBook, repos.Author, adapters.WordAligner, translator, dto.Config{WaitDuration: cfg.WaitMilliseconds, MaxRequestCount: cfg.MaxRequestCount}, notificationService),
	}
}

type wsSender struct{}

func (s *wsSender) SendMessage(userId domain.ID, msg *websocket.Message) {
	websocket.SendMessage(userId, msg)
}

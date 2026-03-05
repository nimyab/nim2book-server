package app

import (
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nimyab/nim2book-back/config"
	_ "github.com/nimyab/nim2book-back/docs"
	"github.com/nimyab/nim2book-back/internal/controller/websocket"
	customMiddleware "github.com/nimyab/nim2book-back/internal/middleware"
	"github.com/nimyab/nim2book-back/internal/services/auth/google_login"
	"github.com/nimyab/nim2book-back/internal/services/auth/login"
	"github.com/nimyab/nim2book-back/internal/services/auth/logout"
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
	"github.com/nimyab/nim2book-back/internal/services/notification"
	"github.com/nimyab/nim2book-back/internal/services/personal_user_book/get_personal_user_book"
	"github.com/nimyab/nim2book-back/internal/services/personal_user_book/get_personal_user_books"
	"github.com/nimyab/nim2book-back/internal/services/personal_user_book/update_personal_user_book"
	"github.com/nimyab/nim2book-back/internal/services/translate/translate_book"
	"github.com/nimyab/nim2book-back/internal/services/translate/translate_personal_book"
	"github.com/nimyab/nim2book-back/internal/services/user/me"
	"github.com/nimyab/nim2book-back/internal/services/user/metadata"
	"github.com/nimyab/nim2book-back/pkg/validator"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// setupHTTPServer initializes the Echo server with middleware and routes
func (a *App) setupHTTPServer(services *Services) {
	e := echo.New()

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(customMiddleware.Logger())

	// CORS configuration
	allowOrigins := []string{
		"https://nim2book.ru",
		"https://www.nim2book.ru",
	}

	if a.config.Env == config.EnvLocal || a.config.Env == config.EnvDev {
		allowOrigins = append(allowOrigins,
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:8080",
		)
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     allowOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch},
		AllowCredentials: true,
	}))

	// Validator
	e.Validator = validator.New()

	// Swagger
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Prometheus metrics
	e.Use(echoprometheus.NewMiddleware("nim2book"))
	e.GET("/metrics", echoprometheus.NewHandler())

	// WebSocket
	e.GET("/ws", websocket.MakeSocketConnHandler(a.config))

	// Setup routes
	if err := a.setupRoutes(e, services); err != nil {
		panic(err)
	}

	a.server = e
}

// setupRoutes defines all HTTP routes
func (a *App) setupRoutes(e *echo.Echo, services *Services) error {
	// JWT middleware
	jwtMiddleware := customMiddleware.JWT(a.config.JWTSecret)
	adminRoleMiddleware := customMiddleware.AdminRole()
	vipRoleMiddleware := customMiddleware.VIPRole()

	apiV1 := e.Group("/api/v1")
	{
		// Health check
		apiV1.GET("/health", func(c echo.Context) error {
			return c.JSON(200, map[string]string{"status": "ok"})
		})

		// Translate routes
		apiV1.POST("/translate/book", translate_book.MakeHTTPv1Handler(services.TranslateBook), jwtMiddleware, adminRoleMiddleware)

		apiV1.POST("/translate/personal-user-book", translate_personal_book.MakeHTTPv1Handler(services.TranslatePersonalBook), jwtMiddleware, vipRoleMiddleware)

		// Auth routes
		apiV1.POST("/auth/register", register.MakeHTTPv1Handler(services.Register))

		apiV1.POST("/auth/login", login.MakeHTTPv1Handler(services.Login, a.config))

		apiV1.POST("/auth/logout", logout.MakeHTTPv1Handler())

		apiV1.POST("/auth/refresh", refresh.MakeHTTPv1Handler(services.Refresh, a.config))

		apiV1.POST("/auth/google", google_login.MakeHTTPv1Handler(services.GoogleLogin, a.config))

		// User routes
		apiV1.GET("/users/me", me.MakeHTTPv1Handler(services.Me), jwtMiddleware)

		apiV1.PUT("/users/metadata", metadata.MakeHTTPv1Handler(services.Metadata), jwtMiddleware)

		// Book routes
		apiV1.GET("/books", get_books.MakeHTTPv1Handler(services.GetBooks))

		apiV1.GET("/books/:id", get_book.MakeHTTPv1Handler(services.GetBook))

		apiV1.PUT("/books/:id", update_book.MakeHTTPv1Handler(services.UpdateBook), jwtMiddleware, adminRoleMiddleware)

		apiV1.GET("/books/:book_id/chapters/:chapter_number", get_chapter.MakeHTTPv1Handler(services.GetChapter))

		// Dictionary routes
		apiV1.POST("/dictionary/lookup", lookup.MakeHTTPv1Handler(services.Lookup))

		// FCM Token routes
		apiV1.POST("/fcm-tokens", add_fcm_token.MakeHTTPv1Handler(services.AddFcmToken), jwtMiddleware)

		apiV1.DELETE("/fcm-tokens/:token", delete_fcm_token.MakeHTTPv1Handler(services.DeleteFcmToken), jwtMiddleware)

		// File routes
		apiV1.GET("/files/*", file_public.MakeHTTPv1Handler(services.FilePublic))

		// Genre routes
		apiV1.GET("/genres", get_genres.MakeHTTPv1Handler(services.GetGenres))

		apiV1.GET("/genres/:id", get_genre.MakeHTTPv1Handler(services.GetGenre))

		apiV1.POST("/genres", create_genre.MakeHTTPv1Handler(services.CreateGenre), jwtMiddleware, adminRoleMiddleware)

		apiV1.DELETE("/genres/:id", delete_genre.MakeHTTPv1Handler(services.DeleteGenre), jwtMiddleware, adminRoleMiddleware)

		// Notification routes
		apiV1.POST("/notifications", notification.MakeHTTPv1Handler(services.Notification), jwtMiddleware, adminRoleMiddleware)

		// Personal User Book routes
		apiV1.GET("/personal-user-books", get_personal_user_books.MakeHTTPv1Handler(services.GetPersonalUserBooks), jwtMiddleware)

		apiV1.GET("/personal-user-books/:id", get_personal_user_book.MakeHTTPv1Handler(services.GetPersonalUserBook), jwtMiddleware)

		apiV1.PUT("/personal-user-books/:id", update_personal_user_book.MakeHTTPv1Handler(services.UpdatePersonalUserBook), jwtMiddleware)
	}

	return nil
}

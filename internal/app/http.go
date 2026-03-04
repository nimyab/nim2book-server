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
	"github.com/samber/do/v2"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// setupHTTPServer initializes the Echo server with middleware and routes
func (a *App) setupHTTPServer() {
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
	if err := a.setupRoutes(e); err != nil {
		panic(err)
	}

	a.server = e
}

// setupRoutes defines all HTTP routes
func (a *App) setupRoutes(e *echo.Echo) error {
	// JWT middleware
	jwtMiddleware := customMiddleware.JWT(a.config.JWTSecret)
	adminRoleMiddleware := customMiddleware.AdminRole()

	apiV1 := e.Group("/api/v1")
	{
		// Health check
		apiV1.GET("/health", func(c echo.Context) error {
			return c.JSON(200, map[string]string{"status": "ok"})
		})

		// Translate routes
		vipRoleMiddleware := customMiddleware.VIPRole()

		svcTranslateBook, err := do.Invoke[*translate_book.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.POST("/translate/book", translate_book.MakeHTTPv1Handler(svcTranslateBook), jwtMiddleware, adminRoleMiddleware)

		svcTranslatePersonalUserBook, err := do.Invoke[*translate_personal_book.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.POST("/translate/personal-user-book", translate_personal_book.MakeHTTPv1Handler(svcTranslatePersonalUserBook), jwtMiddleware, vipRoleMiddleware)

		// Auth routes
		svcRegister, err := do.Invoke[*register.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.POST("/auth/register", register.MakeHTTPv1Handler(svcRegister))

		svcLogin, err := do.Invoke[*login.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.POST("/auth/login", login.MakeHTTPv1Handler(svcLogin, a.config))

		apiV1.POST("/auth/logout", logout.MakeHTTPv1Handler())

		svcRefresh, err := do.Invoke[*refresh.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.POST("/auth/refresh", refresh.MakeHTTPv1Handler(svcRefresh, a.config))

		svcGoogleLogin, err := do.Invoke[*google_login.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.POST("/auth/google", google_login.MakeHTTPv1Handler(svcGoogleLogin, a.config))

		// User routes
		svcMe, err := do.Invoke[*me.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.GET("/users/me", me.MakeHTTPv1Handler(svcMe), jwtMiddleware)

		svcMetadata, err := do.Invoke[*metadata.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.PUT("/users/metadata", metadata.MakeHTTPv1Handler(svcMetadata), jwtMiddleware)

		// Book routes
		svcGetBooks, err := do.Invoke[*get_books.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.GET("/books", get_books.MakeHTTPv1Handler(svcGetBooks))

		svcGetBook, err := do.Invoke[*get_book.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.GET("/books/:id", get_book.MakeHTTPv1Handler(svcGetBook))

		svcUpdateBook, err := do.Invoke[*update_book.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.PUT("/books/:id", update_book.MakeHTTPv1Handler(svcUpdateBook), jwtMiddleware, adminRoleMiddleware)

		svcGetChapter, err := do.Invoke[*get_chapter.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.GET("/books/:book_id/chapters/:chapter_number", get_chapter.MakeHTTPv1Handler(svcGetChapter))

		// Dictionary routes
		svcLookup, err := do.Invoke[*lookup.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.POST("/dictionary/lookup", lookup.MakeHTTPv1Handler(svcLookup), jwtMiddleware)

		// FCM Token routes
		svcAddFcmToken, err := do.Invoke[*add_fcm_token.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.POST("/fcm-tokens", add_fcm_token.MakeHTTPv1Handler(svcAddFcmToken), jwtMiddleware)

		svcDeleteFcmToken, err := do.Invoke[*delete_fcm_token.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.DELETE("/fcm-tokens/:token", delete_fcm_token.MakeHTTPv1Handler(svcDeleteFcmToken), jwtMiddleware)

		// File routes
		svcFilePublic, err := do.Invoke[*file_public.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.GET("/files/*", file_public.MakeHTTPv1Handler(svcFilePublic))

		// Genre routes
		svcGetGenres, err := do.Invoke[*get_genres.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.GET("/genres", get_genres.MakeHTTPv1Handler(svcGetGenres))

		svcGetGenre, err := do.Invoke[*get_genre.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.GET("/genres/:id", get_genre.MakeHTTPv1Handler(svcGetGenre))

		svcCreateGenre, err := do.Invoke[*create_genre.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.POST("/genres", create_genre.MakeHTTPv1Handler(svcCreateGenre), jwtMiddleware, adminRoleMiddleware)

		svcDeleteGenre, err := do.Invoke[*delete_genre.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.DELETE("/genres/:id", delete_genre.MakeHTTPv1Handler(svcDeleteGenre), jwtMiddleware, adminRoleMiddleware)

		// Notification routes
		svcNotification, err := do.Invoke[*notification.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.POST("/notifications", notification.MakeHTTPv1Handler(svcNotification), jwtMiddleware, adminRoleMiddleware)

		// Personal User Book routes
		svcGetPersonalUserBooks, err := do.Invoke[*get_personal_user_books.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.GET("/personal-user-books", get_personal_user_books.MakeHTTPv1Handler(svcGetPersonalUserBooks), jwtMiddleware)

		svcGetPersonalUserBook, err := do.Invoke[*get_personal_user_book.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.GET("/personal-user-books/:id", get_personal_user_book.MakeHTTPv1Handler(svcGetPersonalUserBook), jwtMiddleware)

		svcUpdatePersonalUserBook, err := do.Invoke[*update_personal_user_book.Service](a.injector)
		if err != nil {
			return err
		}
		apiV1.PUT("/personal-user-books/:id", update_personal_user_book.MakeHTTPv1Handler(svcUpdatePersonalUserBook), jwtMiddleware)
	}

	return nil
}

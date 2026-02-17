package app

import (
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	"github.com/nimyab/nim2book-back/internal/services/genre/update_genre"
	"github.com/nimyab/nim2book-back/internal/services/notification"
	"github.com/nimyab/nim2book-back/internal/services/personal_user_book/get_personal_user_book"
	"github.com/nimyab/nim2book-back/internal/services/personal_user_book/get_personal_user_books"
	"github.com/nimyab/nim2book-back/internal/services/personal_user_book/update_personal_user_book"
	"github.com/nimyab/nim2book-back/internal/services/translate/translate_book"
	"github.com/nimyab/nim2book-back/internal/services/translate/translate_personal_user_book"
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
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://localhost*",
			"http://*.nim2book.ru",
			"http://nim2book.ru",
			"https://*.nim2book.ru",
			"https://nim2book.ru",
		},
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
	a.setupRoutes(e)

	a.server = e
}

// setupRoutes defines all HTTP routes
func (a *App) setupRoutes(e *echo.Echo) {
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
		if svc, err := do.Invoke[*translate_book.Service](a.injector); err == nil {
			apiV1.POST("/translate/book", translate_book.MakeHTTPv1Handler(svc), jwtMiddleware, adminRoleMiddleware)
		}
		if svc, err := do.Invoke[*translate_personal_user_book.Service](a.injector); err == nil {
			apiV1.POST("/translate/personal-user-book", translate_personal_user_book.MakeHTTPv1Handler(svc), jwtMiddleware, vipRoleMiddleware)
		}

		// Book routes
		if svc, err := do.Invoke[*get_chapter.Service](a.injector); err == nil {
			apiV1.GET("/book/get-chapter/:path", get_chapter.MakeHTTPv1Handler(svc))
		}
		if svc, err := do.Invoke[*get_books.Service](a.injector); err == nil {
			apiV1.GET("/book", get_books.MakeHTTPv1Handler(svc))
		}
		if svc, err := do.Invoke[*get_book.Service](a.injector); err == nil {
			apiV1.GET("/book/:id", get_book.MakeHTTPv1Handler(svc))
		}
		if svc, err := do.Invoke[*update_book.Service](a.injector); err == nil {
			apiV1.PUT("/book/:id", update_book.MakeHTTPv1Handler(svc), jwtMiddleware, adminRoleMiddleware)
		}

		// Personal user book routes
		if svc, err := do.Invoke[*get_personal_user_books.Service](a.injector); err == nil {
			apiV1.GET("/personal-user-book", get_personal_user_books.MakeHTTPv1Handler(svc), jwtMiddleware)
		}
		if svc, err := do.Invoke[*get_personal_user_book.Service](a.injector); err == nil {
			apiV1.GET("/personal-user-book/:id", get_personal_user_book.MakeHTTPv1Handler(svc), jwtMiddleware)
		}
		if svc, err := do.Invoke[*update_personal_user_book.Service](a.injector); err == nil {
			apiV1.PUT("/personal-user-book/:id", update_personal_user_book.MakeHTTPv1Handler(svc), jwtMiddleware)
		}

		// Auth routes
		if svc, err := do.Invoke[*register.Service](a.injector); err == nil {
			apiV1.POST("/auth/register", register.MakeHTTPv1Handler(svc))
		}
		if svc, err := do.Invoke[*login.Service](a.injector); err == nil {
			apiV1.POST("/auth/login", login.MakeHTTPv1Handler(svc, a.config))
		}
		if svc, err := do.Invoke[*google_login.Service](a.injector); err == nil {
			apiV1.POST("/auth/google-login", google_login.MakeHTTPv1Handler(svc, a.config))
		}
		apiV1.POST("/auth/logout", logout.MakeHTTPv1Handler())
		if svc, err := do.Invoke[*refresh.Service](a.injector); err == nil {
			apiV1.POST("/auth/refresh", refresh.MakeHTTPv1Handler(svc, a.config))
		}

		// User routes
		if svc, err := do.Invoke[*me.Service](a.injector); err == nil {
			apiV1.GET("/user/me", me.MakeHTTPv1Handler(svc), jwtMiddleware)
		}
		if svc, err := do.Invoke[*metadata.Service](a.injector); err == nil {
			apiV1.PUT("/user/metadata", metadata.MakeHTTPv1Handler(svc), jwtMiddleware)
		}

		// Dictionary routes
		if svc, err := do.Invoke[*lookup.Service](a.injector); err == nil {
			apiV1.POST("/dictionary/lookup", lookup.MakeHTTPv1Handler(svc))
		}

		// File routes
		if svc, err := do.Invoke[*file_public.Service](a.injector); err == nil {
			apiV1.GET("/file/public", file_public.MakeHTTPv1Handler(svc))
		}

		// FCM token routes
		if svc, err := do.Invoke[*add_fcm_token.Service](a.injector); err == nil {
			apiV1.POST("/fcm-token/add", add_fcm_token.MakeHTTPv1Handler(svc), jwtMiddleware)
		}
		if svc, err := do.Invoke[*delete_fcm_token.Service](a.injector); err == nil {
			apiV1.DELETE("/fcm-token/delete", delete_fcm_token.MakeHTTPv1Handler(svc), jwtMiddleware)
		}

		// Notification routes
		if svc, err := do.Invoke[*notification.Service](a.injector); err == nil {
			apiV1.POST("/notification/test", notification.MakeHTTPv1Handler(svc), jwtMiddleware)
		}

		// Genre routes
		if svc, err := do.Invoke[*get_genres.Service](a.injector); err == nil {
			apiV1.GET("/genre", get_genres.MakeHTTPv1Handler(svc))
		}
		if svc, err := do.Invoke[*get_genre.Service](a.injector); err == nil {
			apiV1.GET("/genre/:id", get_genre.MakeHTTPv1Handler(svc))
		}
		if svc, err := do.Invoke[*create_genre.Service](a.injector); err == nil {
			apiV1.POST("/genre", create_genre.MakeHTTPv1Handler(svc), jwtMiddleware, adminRoleMiddleware)
		}
		if svc, err := do.Invoke[*update_genre.Service](a.injector); err == nil {
			apiV1.PUT("/genre/:id", update_genre.MakeHTTPv1Handler(svc), jwtMiddleware, adminRoleMiddleware)
		}
		if svc, err := do.Invoke[*delete_genre.Service](a.injector); err == nil {
			apiV1.DELETE("/genre/:id", delete_genre.MakeHTTPv1Handler(svc), jwtMiddleware, adminRoleMiddleware)
		}
	}
}

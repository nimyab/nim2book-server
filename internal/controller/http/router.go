package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/nimyab/nim2book-back/docs"
	"github.com/nimyab/nim2book-back/internal/auth/login"
	"github.com/nimyab/nim2book-back/internal/auth/logout"
	"github.com/nimyab/nim2book-back/internal/auth/refresh"
	"github.com/nimyab/nim2book-back/internal/auth/register"
	"github.com/nimyab/nim2book-back/internal/book/get_book"
	"github.com/nimyab/nim2book-back/internal/book/get_books"
	"github.com/nimyab/nim2book-back/internal/book/get_chapter"
	"github.com/nimyab/nim2book-back/internal/book/update_book"
	"github.com/nimyab/nim2book-back/internal/controller/websocket"
	"github.com/nimyab/nim2book-back/internal/dictionary/lookup"
	"github.com/nimyab/nim2book-back/internal/file/file_public"
	customMiddleware "github.com/nimyab/nim2book-back/internal/middleware"
	"github.com/nimyab/nim2book-back/internal/translate/translate_book"
	"github.com/nimyab/nim2book-back/internal/user/me"
	"github.com/nimyab/nim2book-back/pkg/validator"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func Router(secretKey string) *echo.Echo {
	e := echo.New()

	jwtMiddleware := customMiddleware.JWT(secretKey)
	adminRoleMiddleware := customMiddleware.AdminRole()

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	e.Use(customMiddleware.Logger())

	e.Validator = validator.New()

	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/ws", websocket.NewSocketConn)

	apiV1 := e.Group("/api/v1")
	{
		apiV1.POST("/translate/book", translate_book.HTTPv1, jwtMiddleware)

		apiV1.GET("/book/get-chapter/:path", get_chapter.HTTPv1)
		apiV1.GET("/book", get_books.HTTPv1)
		apiV1.GET("/book/:id", get_book.HTTPv1)
		apiV1.PUT("/book/:id", update_book.HTTPv1, jwtMiddleware, adminRoleMiddleware)

		apiV1.POST("/auth/register", register.HTTPv1)
		apiV1.POST("/auth/login", login.HTTPv1)
		apiV1.POST("/auth/logout", logout.HTTPv1)
		apiV1.POST("/auth/refresh", refresh.HTTPv1)

		apiV1.GET("/user/me", me.HTTPv1, jwtMiddleware)

		apiV1.POST("/dictionary/lookup", lookup.HTTPv1)

		apiV1.GET("/file/public", file_public.HTTPv1)
	}

	return e
}

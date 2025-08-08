package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/nimyab/nim2book-back/internal/auth/login"
	"github.com/nimyab/nim2book-back/internal/auth/logout"
	"github.com/nimyab/nim2book-back/internal/auth/register"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/nimyab/nim2book-back/docs"
	"github.com/nimyab/nim2book-back/internal/book/get_book"
	"github.com/nimyab/nim2book-back/internal/book/get_books"
	"github.com/nimyab/nim2book-back/internal/book/get_chapter"
	customMiddleware "github.com/nimyab/nim2book-back/internal/middleware"
	"github.com/nimyab/nim2book-back/internal/translate/translate_book"
	"github.com/nimyab/nim2book-back/pkg/validator"
)

func Router() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	e.Use(customMiddleware.Logger())

	e.Validator = validator.New()

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	apiV1 := e.Group("/api/v1")
	{
		apiV1.POST("/translate/book", translate_book.HTTPv1)

		apiV1.GET("/book/get-chapter/:path", get_chapter.HTTPv1)
		apiV1.GET("/book", get_books.HTTPv1)
		apiV1.GET("/book/:id", get_book.HTTPv1)

		apiV1.POST("/auth/register", register.HTTPv1)
		apiV1.POST("/auth/login", login.HTTPv1)
		apiV1.POST("/auth/logout", logout.HTTPv1)
	}

	return e
}

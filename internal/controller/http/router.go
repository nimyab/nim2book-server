package http

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"log/slog"
	"regexp"
	"time"

	_ "github.com/nimyab/nim2book-back/docs"
	"github.com/nimyab/nim2book-back/internal/book/get_book"
	"github.com/nimyab/nim2book-back/internal/book/get_books"
	"github.com/nimyab/nim2book-back/internal/book/get_chapter"
	"github.com/nimyab/nim2book-back/internal/translate/translate_book"
	"github.com/nimyab/nim2book-back/pkg/validator"
)

func Router() *echo.Echo {
	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.RequestID())
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if regexp.MustCompile(`(?i).*swagger.*`).MatchString(c.Request().URL.Path) {
				return next(c)
			}

			slog.Info(
				"Request received",
				slog.String("method", c.Request().Method),
				slog.String("path", c.Request().URL.Path),
				slog.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
				slog.String("user_agent", c.Request().UserAgent()),
				slog.String("host", c.Request().Host),
			)

			t1 := time.Now()
			defer func() {
				slog.Info(
					"Request completed",
					slog.String("method", c.Request().Method),
					slog.String("path", c.Request().URL.Path),
					slog.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)),
					slog.String("user_agent", c.Request().UserAgent()),
					slog.String("host", c.Request().Host),
					slog.Int("status", c.Response().Status),
					slog.String("duration", time.Since(t1).String()),
				)
			}()

			return next(c)
		}
	})

	e.Validator = validator.New()

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	apiV1 := e.Group("/api/v1")
	{
		apiV1.POST("/translate/book", translate_book.HTTPv1)

		apiV1.GET("/book/get-chapter/:path", get_chapter.HTTPv1)
		apiV1.GET("/book", get_books.HTTPv1)
		apiV1.GET("/book/:id", get_book.HTTPv1)
	}

	return e
}

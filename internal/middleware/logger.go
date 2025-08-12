package middleware

import (
	"github.com/labstack/echo/v4"
	"log/slog"
	"regexp"
	"time"
)

func Logger() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
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
				slog.String("remote_ip", c.Request().RemoteAddr),
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
					slog.String("remote_ip", c.Request().RemoteAddr),
					slog.Int("status", c.Response().Status),
					slog.String("duration", time.Since(t1).String()),
				)
			}()

			return next(c)
		}
	}
}

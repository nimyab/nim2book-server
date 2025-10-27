package middleware

import (
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/pkg/jwt"
)

func VIPRole() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			payload := jwt.GetUserPayload(c)
			slog.Info("payload", slog.Any("payload", payload))

			if !payload.IsVIP {
				return echo.NewHTTPError(http.StatusForbidden)
			}
			return next(c)
		}
	}
}

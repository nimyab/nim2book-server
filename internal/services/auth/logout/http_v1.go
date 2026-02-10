package logout

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// MakeHTTPv1Handler godoc
// @Summary	Logout user
// @Tags	auth
// @Produce	json
// @Success	200		{object}	Output
// @Router	/auth/logout	[post]
func MakeHTTPv1Handler() echo.HandlerFunc {
	return func(c echo.Context) error {
		cookie := &http.Cookie{
			HttpOnly: true,
			Name:     "refresh_token",
			Value:    "",
			Expires:  time.Now().Add(-24 * time.Hour),
		}
		c.SetCookie(cookie)
		return c.JSON(http.StatusOK, &Output{
			Success: true,
		})
	}
}

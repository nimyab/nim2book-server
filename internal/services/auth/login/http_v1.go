package login

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/config"
)

// MakeHTTPv1Handler creates HTTP handler with dependencies
func MakeHTTPv1Handler(svc *Service, cfg *config.Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		input := new(Input)

		if err := c.Bind(input); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}

		if err := c.Validate(input); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}

		output, err := svc.Login(input)
		if errors.Is(ErrInternal, err) {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		} else if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}

		cookie := &http.Cookie{
			HttpOnly: true,
			Name:     "refresh_token",
			Value:    output.RefreshToken,
			Expires:  time.Now().Add(cfg.JWTRefreshTime),
		}
		c.SetCookie(cookie)
		return c.JSON(http.StatusOK, output)
	}
}

// HTTPv1 godoc
// @Summary	Login user
// @Tags	auth
// @Accept  json
// @Produce	json
// @Param	data	body	Input	true	"body"
// @Success	200		{object}	Output
// @Router	/auth/login	[post]
// Deprecated: Use MakeHTTPv1Handler instead
func HTTPv1(c echo.Context) error {
	panic("HTTPv1 is deprecated, use MakeHTTPv1Handler instead")
}

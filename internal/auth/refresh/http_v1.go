package refresh

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/config"
)

// HTTPv1 godoc
// @Summary	Refresh user
// @Tags	auth
// @Accept  json
// @Produce	json
// @Param	data	body	Input	false	"body"
// @Success	200		{object}	Output
// @Router	/auth/refresh	[post]
func HTTPv1(c echo.Context) error {
	input := new(Input)

	if err := c.Bind(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	token, _ := c.Cookie("refresh_token")
	if input.RefreshToken == "" && token != nil {
		input.RefreshToken = token.Value
	}

	if err := c.Validate(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	output, err := service.Refresh(input)
	if errors.Is(err, ErrInternal) {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	} else if errors.Is(err, ErrParseTokenFailed) {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": err.Error()})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	cookie := &http.Cookie{
		HttpOnly: true,
		Name:     "refresh_token",
		Value:    output.RefreshToken,
		Expires:  time.Now().Add(config.GetConfig().JWTRefreshTime),
	}
	c.SetCookie(cookie)
	return c.JSON(http.StatusOK, output)
}

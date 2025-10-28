package login

import (
	"errors"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/config"
)

// HTTPv1 godoc
// @Summary	Login user
// @Tags	auth
// @Accept  json
// @Produce	json
// @Param	data	body	Input	true	"body"
// @Success	200		{object}	Output
// @Router	/auth/login	[post]
func HTTPv1(c echo.Context) error {
	input := new(Input)

	if err := c.Bind(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if err := c.Validate(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	output, err := service.Login(input)
	if errors.Is(ErrInternal, err) {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	} else if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
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

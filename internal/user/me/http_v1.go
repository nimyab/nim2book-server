package me

import (
	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/pkg/jwt"
	"net/http"
)

// HTTPv1 godoc
// @Summary	Get user info
// @Tags	user
// @Produce	application/json
// @Security BearerAuth
// @Success	200		{object}	Output
// @Router	/user/me	[get]
func HTTPv1(c echo.Context) error {
	payload := jwt.GetUserPayload(c)
	input := &Input{
		UserId: payload.Id,
	}

	if err := c.Validate(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": err.Error(),
		})
	}

	output, err := service.Me(input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, output)
}

package register

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
)

// HTTPv1 godoc
// @Summary	Register user
// @Tags	auth
// @Accept  json
// @Produce	json
// @Param	data	body	Input	true	"body"
// @Success	200		{object}	Output
// @Router	/auth/register	[post]
func HTTPv1(c echo.Context) error {
	input := new(Input)
	if err := c.Bind(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid input",
		})
	}

	if err := c.Validate(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	output, err := service.Register(input)
	if errors.Is(err, ErrInternal) {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	} else if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, output)
}

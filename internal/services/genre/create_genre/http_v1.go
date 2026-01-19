package create_genre

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HTTPv1 godoc
// @Summary	Create genre
// @Tags	genre
// @Security BearerAuth
// @Accept	application/json
// @Produce	application/json
// @Param	input	body	Input	true	"Genre data"
// @Success	201		{object}	Output
// @Router	/genre	[post]
func HTTPv1(c echo.Context) error {
	input := new(Input)
	if err := c.Bind(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid input",
		})
	}

	if err := c.Validate(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": err.Error(),
		})
	}

	output, err := service.CreateGenre(input)
	if errors.Is(err, ErrGenreAlreadyExists) {
		return c.JSON(http.StatusConflict, echo.Map{
			"error": err.Error(),
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, output)
}

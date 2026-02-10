package update_genre

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

// MakeHTTPv1Handler creates HTTP handler with dependencies
func MakeHTTPv1Handler(svc *Service) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		output, err := svc.UpdateGenre(input)
		if errors.Is(err, ErrGenreNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{
				"error": err.Error(),
			})
		}
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

		return c.JSON(http.StatusOK, output)
	}
}

// HTTPv1 godoc
// @Summary	Update genre
// @Tags	genre
// @Security BearerAuth
// @Accept	application/json
// @Produce	application/json
// @Param	id		path	string	true	"Genre id"
// @Param	input	body	Input	true	"Genre data"
// @Success	200		{object}	Output
// @Router	/genre/{id}	[put]
// Deprecated: Use MakeHTTPv1Handler instead
func HTTPv1(c echo.Context) error {
	panic("HTTPv1 is deprecated, use MakeHTTPv1Handler instead")
}

package delete_genre

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// MakeHTTPv1Handler godoc
// @Summary	Delete genre
// @Tags	genre
// @Security BearerAuth
// @Produce	application/json
// @Param	id	path	string	true	"Genre id"
// @Success	200		{object}	Output
// @Router	/genres/{id}	[delete]
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

		output, err := svc.DeleteGenre(c.Request().Context(), input)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, output)
	}
}

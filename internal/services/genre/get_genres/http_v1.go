package get_genres

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HTTPv1 godoc
// @Summary	Get all genres
// @Tags	genre
// @Produce	application/json
// @Success	200		{object}	Output
// @Router	/genre	[get]
func HTTPv1(c echo.Context) error {
	output, err := service.GetGenres()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, output)
}

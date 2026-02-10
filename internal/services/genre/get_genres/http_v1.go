package get_genres

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// MakeHTTPv1Handler creates HTTP handler with dependencies
func MakeHTTPv1Handler(svc *Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		output, err := svc.GetGenres()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, output)
	}
}

// HTTPv1 godoc
// @Summary	Get all genres
// @Tags	genre
// @Produce	application/json
// @Success	200		{object}	Output
// @Router	/genre	[get]
// Deprecated: Use MakeHTTPv1Handler instead
func HTTPv1(c echo.Context) error {
	panic("HTTPv1 is deprecated, use MakeHTTPv1Handler instead")
}

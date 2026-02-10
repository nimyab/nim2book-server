package get_books

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// MakeHTTPv1Handler godoc
// @Summary	Get books
// @Tags	book
// @Produce	application/json
// @Param	author	query	string	false	"Author filter"
// @Param	title	query	string	false	"Title filter"
// @Param	genreId	query	string	false	"Genre ID filter"
// @Param	page	query	int		true	"Page number"
// @Success	200		{object}	Output
// @Router	/book	[get]
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

		output, err := svc.GetBooks(input)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, output)
	}
}

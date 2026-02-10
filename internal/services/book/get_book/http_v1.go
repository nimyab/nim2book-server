package get_book

import (
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

		output, err := svc.GetBook(input)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, output)
	}
}

// HTTPv1 godoc
// @Summary	Get book by id
// @Tags	book
// @Produce	application/json
// @Param	id	path	string	true	"Book id"
// @Success	200		{object}	Output
// @Router	/book/{id}	[get]
// Deprecated: Use MakeHTTPv1Handler instead
func HTTPv1(c echo.Context) error {
	panic("HTTPv1 is deprecated, use MakeHTTPv1Handler instead")
}

package get_books

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HTTPv1 godoc
// @Summary	Get books
// @Tags	book
// @Produce	application/json
// @Param	author	query	string	false	"Chapter path"
// @Param	title	query	string	false	"Chapter path"
// @Param	page	query	string	true	"Chapter path"
// @Success	200		{object}	Output
// @Router	/book	[get]
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

	output, err := service.GetBooks(input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, output)
}

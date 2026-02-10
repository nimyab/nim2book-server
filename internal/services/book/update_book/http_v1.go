package update_book

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// MakeHTTPv1Handler godoc
// @Summary	Update book
// @Tags	book
// @Security BearerAuth
// @Produce	application/json
// @Param	cover	formData	file	false	"Cover file"
// @Param	author	formData	string	false	"Author book"
// @Param	title	formData	string	false	"Title book"
// @Param	id		path	string	true	"Book id"
// @Success	200		{object}	Output
// @Router	/book/{id}	[put]
func MakeHTTPv1Handler(svc *Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		cover, err := c.FormFile("cover")
		if err != nil {
			return c.JSON(http.StatusBadRequest, err.Error())
		}

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

		output, err := svc.UpdateBook(input, cover)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, output)
	}
}

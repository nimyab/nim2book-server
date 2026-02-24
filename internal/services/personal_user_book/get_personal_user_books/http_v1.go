package get_personal_user_books

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/pkg/jwt"
)

// MakeHTTPv1Handler creates HTTP handler with dependencies
// @Summary	Get personal user books
// @Tags	personal-user-book
// @Security BearerAuth
// @Produce	json
// @Param	author	query	string	false	"Author"
// @Param	title	query	string	false	"Title"
// @Param	genreId	query	string	false	"Genre ID"
// @Param	page	query	int		true	"Page"
// @Success	200		{object}	Output
// @Router	/personal-user-books	[get]
func MakeHTTPv1Handler(svc *Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		payload := jwt.GetUserPayload(c)

		input := new(Input)
		if err := c.Bind(input); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "invalid input",
			})
		}

		input.UserId = payload.ID

		if err := c.Validate(input); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": err.Error(),
			})
		}

		output, err := svc.GetPersonalUserBooks(c.Request().Context(), input)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, output)
	}
}

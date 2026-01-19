package get_personal_user_books

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/pkg/jwt"
)

// HTTPv1 godoc
// @Summary	Get personal user books
// @Tags	personal_user_book
// @Security	BearerAuth
// @Produce	application/json
// @Param	author	query	string	false	"Author filter"
// @Param	title	query	string	false	"Title filter"
// @Param	genreId	query	string	false	"Genre ID filter"
// @Param	page	query	int		true	"Page number"
// @Success	200		{object}	Output
// @Failure	400		{object}	map[string]string
// @Failure	401		{object}	map[string]string
// @Failure	500		{object}	map[string]string
// @Router	/personal-user-book	[get]
func HTTPv1(c echo.Context) error {
	payload := jwt.GetUserPayload(c)

	input := new(Input)
	if err := c.Bind(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid input",
		})
	}

	input.UserId = payload.Id

	if err := c.Validate(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": err.Error(),
		})
	}

	output, err := service.GetPersonalUserBooks(input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, output)
}

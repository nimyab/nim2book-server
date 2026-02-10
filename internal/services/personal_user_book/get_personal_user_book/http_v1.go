package get_personal_user_book

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/pkg/jwt"
)

// MakeHTTPv1Handler creates HTTP handler with dependencies
func MakeHTTPv1Handler(svc *Service) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		output, err := svc.GetPersonalUserBook(input)
		if err != nil {
			if errors.Is(err, ErrBookNotFound) {
				return c.JSON(http.StatusNotFound, echo.Map{
					"error": "book not found",
				})
			}
			if errors.Is(err, ErrForbidden) {
				return c.JSON(http.StatusForbidden, echo.Map{
					"error": "you don't have access to this book",
				})
			}
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, output)
	}
}

// HTTPv1 godoc
// @Summary	Get personal user book by ID
// @Tags	personal_user_book
// @Security	BearerAuth
// @Produce	application/json
// @Param	id	path	string	true	"Book ID"
// @Success	200	{object}	Output
// @Failure	400	{object}	map[string]string
// @Failure	401	{object}	map[string]string
// @Failure	403	{object}	map[string]string
// @Failure	404	{object}	map[string]string
// @Failure	500	{object}	map[string]string
// @Router	/personal-user-book/{id}	[get]
// Deprecated: Use MakeHTTPv1Handler instead
func HTTPv1(c echo.Context) error {
	panic("HTTPv1 is deprecated, use MakeHTTPv1Handler instead")
}

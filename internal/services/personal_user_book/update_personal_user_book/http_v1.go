package update_personal_user_book

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

		cover, _ := c.FormFile("cover")

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

		output, err := svc.UpdatePersonalUserBook(input, cover)
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
			if errors.Is(err, ErrFailedToOpenFile) || errors.Is(err, ErrFailedToReadFile) {
				return c.JSON(http.StatusBadRequest, echo.Map{
					"error": "invalid file",
				})
			}
			if errors.Is(err, ErrFailedToUploadFile) {
				return c.JSON(http.StatusInternalServerError, echo.Map{
					"error": "failed to upload file",
				})
			}
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, output)
	}
}

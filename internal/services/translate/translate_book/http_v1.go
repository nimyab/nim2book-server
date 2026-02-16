package translate_book

import (
	"net/http"

	"github.com/nimyab/nim2book-back/pkg/jwt"

	"github.com/labstack/echo/v4"
)

// MakeHTTPv1Handler creates HTTP handler with dependencies
func MakeHTTPv1Handler(svc *Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		userPayload := jwt.GetUserPayload(c)

		bookFile, err := c.FormFile("file")
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "file is required",
			})
		}
		input := new(Input)
		if err = c.Bind(input); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "invalid input",
			})
		}
		if err = c.Validate(input); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": err.Error(),
			})
		}

		output, err := svc.TranslateBook(c.Request().Context(), input, bookFile, userPayload.ID)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusCreated, output)
	}
}

package me

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/pkg/jwt"
)

// MakeHTTPv1Handler creates HTTP handler with dependencies
func MakeHTTPv1Handler(svc *Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		payload := jwt.GetUserPayload(c)
		input := &Input{
			UserId: payload.Id,
		}

		if err := c.Validate(input); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": err.Error(),
			})
		}

		output, err := svc.Me(input)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{
				"error": err.Error(),
			})
		}

		return c.JSON(http.StatusOK, output)
	}
}

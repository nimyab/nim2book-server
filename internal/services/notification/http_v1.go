package notification

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/internal/domain"
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

		if err := c.Validate(input); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": err.Error(),
			})
		}

		svc.ProcessNotification(context.Background(), &domain.Notification{
			UserId: payload.Id,
			Type:   domain.NotificationTest,
			Data: &domain.NotificationTestData{
				Title: input.Title,
				Body:  input.Body,
			},
		})

		return c.String(http.StatusOK, "OK")
	}
}

// HTTPv1 godoc
// @Summary	Notification test
// @Tags	notification
// @Security BearerAuth
// @Accept  json
// @Produce	json
// @Param	data	body	Input	true	"body"
// @Success	200		{string}	string
// @Router	/notification/test	[post]
// Deprecated: Use MakeHTTPv1Handler instead
func HTTPv1(c echo.Context) error {
	panic("HTTPv1 is deprecated, use MakeHTTPv1Handler instead")
}

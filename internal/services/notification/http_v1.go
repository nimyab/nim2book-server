package notification

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/pkg/jwt"
)

// HTTPv1 godoc
// @Summary	Notification test
// @Tags	notification
// @Security BearerAuth
// @Accept  json
// @Produce	json
// @Param	data	body	Input	true	"body"
// @Success	200		{string}	string
// @Router	/notification/test	[post]
func HTTPv1(c echo.Context) error {
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

	service.ProcessNotification(context.Background(), models.Notification{
		UserId: payload.Id,
		Type:   models.NotificationTest,
		Data: &models.NotificationTestData{
			Title: input.Title,
			Body:  input.Body,
		},
	})

	return c.String(http.StatusOK, "OK")
}

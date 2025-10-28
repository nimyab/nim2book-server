package add_fcm_token

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/pkg/jwt"
)

// HTTPv1 godoc
// @Summary	Add fcm token for send notifications
// @Security BearerAuth
// @Tags	fcm-token
// @Accept  json
// @Produce	json
// @Param	data	body	Input	true	"body"
// @Success	200		{object}	Output
// @Router	/fcm-token/add	[post]
func HTTPv1(c echo.Context) error {
	userId := jwt.GetUserPayload(c).Id

	input := new(Input)
	if err := c.Bind(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if err := c.Validate(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	output, err := service.AddFcmToken(input, userId)
	if errors.Is(err, ErrInternal) {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	} else if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, output)
}

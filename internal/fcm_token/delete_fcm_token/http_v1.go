package delete_fcm_token

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/pkg/jwt"
)

// HTTPv1 godoc
// @Summary	Delete fcm token
// @Security BearerAuth
// @Tags	fcm-token
// @Produce	json
// @Param	token	query	string	true	"fcm token"
// @Success	200		{object}	Output
// @Router	/fcm-token/delete	[delete]
func HTTPv1(c echo.Context) error {
	userId := jwt.GetUserPayload(c).Id

	input := new(Input)
	if err := c.Bind(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if err := c.Validate(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	output, err := service.DeleteFcmToken(input, userId)
	if errors.Is(err, ErrInternal) {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	} else if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, output)
}

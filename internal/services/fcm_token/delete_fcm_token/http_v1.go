package delete_fcm_token

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/pkg/jwt"
)

// MakeHTTPv1Handler creates HTTP handler with dependencies
// @Summary	Delete FCM token
// @Tags	fcm-token
// @Security BearerAuth
// @Produce	json
// @Param	token	query	string	true	"FCM token"
// @Success	200		{object}	Output
// @Router	/fcm-tokens/{token}	[delete]
func MakeHTTPv1Handler(svc *Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		userId := jwt.GetUserPayload(c).ID

		input := new(Input)
		if err := c.Bind(input); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}

		if err := c.Validate(input); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}

		output, err := svc.DeleteFcmToken(c.Request().Context(), input, userId)
		if errors.Is(err, ErrInternal) {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
		} else if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, output)
	}
}

package metadata

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/nimyab/nim2book-back/pkg/jwt"
)

// HTTPv1 godoc
// @Summary	Update user metadata
// @Tags	user
// @Produce	application/json
// @Security BearerAuth
// @Param	input	body		Input	true	"Input data"
// @Success	200		{object}	Output
// @Router	/user/metadata	[put]
func HTTPv1(c echo.Context) error {
	payload := jwt.GetUserPayload(c)
	input := &Input{
		UserId: payload.Id,
	}

	if err := c.Bind(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": err.Error(),
		})
	}

	if err := c.Validate(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": err.Error(),
		})
	}

	output, err := service.UpdateMetadata(input)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, output)
}

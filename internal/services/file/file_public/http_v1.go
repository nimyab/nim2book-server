package file_public

import (
	"bytes"
	"net/http"

	"github.com/labstack/echo/v4"
)

// MakeHTTPv1Handler godoc
// @Summary	Get file data
// @Tags	file
// @Param	path	query	string	true	"path"
// @Success	200		{file} 	binary
// @Router	/file/public	[get]
func MakeHTTPv1Handler(svc *Service) echo.HandlerFunc {
	return func(c echo.Context) error {
		var input = new(Input)
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

		data, err := svc.GetFile(input)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": err.Error(),
			})
		}

		return c.Stream(http.StatusOK, http.DetectContentType(data), bytes.NewReader(data))
	}
}

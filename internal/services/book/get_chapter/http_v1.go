package get_chapter

import (
	"bytes"
	"net/http"

	"github.com/labstack/echo/v4"
)

// MakeHTTPv1Handler creates HTTP handler with dependencies
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

		data, err := svc.GetChapter(input)
		if err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "failed to fetch chapter",
			})
		}

		return c.Stream(http.StatusOK, "application/json", bytes.NewReader(data))
	}
}

// HTTPv1 godoc
// @Summary	Get book chapter
// @Tags	book
// @Produce	application/json
// @Param	path	path	string	true	"Chapter path"
// @Success	200		{object}	domain.ChapterAlignNode
// @Router	/book/get-chapter/{path}	[get]
// Deprecated: Use MakeHTTPv1Handler instead
func HTTPv1(c echo.Context) error {
	panic("HTTPv1 is deprecated, use MakeHTTPv1Handler instead")
}

package get_chapter

import (
	"bytes"
	"net/http"

	"github.com/labstack/echo/v4"
)

// MakeHTTPv1Handler creates HTTP handler with dependencies
// @Summary	Get chapter content
// @Tags	book
// @Produce	application/json
// @Param	path	query	string	true	"Chapter path"
// @Success	200		{string}	string	"Chapter content"
// @Router	/books/{book_id}/chapters/{chapter_number}	[get]
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

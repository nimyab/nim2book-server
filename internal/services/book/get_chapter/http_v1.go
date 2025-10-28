package get_chapter

import (
	"bytes"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HTTPv1 godoc
// @Summary	Get book chapter
// @Tags	book
// @Produce	application/json
// @Param	path	path	string	true	"Chapter path"
// @Success	200		{object}	domain.ChapterAlignNode
// @Router	/book/get-chapter/{path}	[get]
func HTTPv1(c echo.Context) error {
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

	data, err := service.GetChapter(input)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "failed to fetch chapter",
		})
	}

	return c.Stream(http.StatusOK, "application/json", bytes.NewReader(data))
}

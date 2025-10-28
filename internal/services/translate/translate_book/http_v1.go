package translate_book

import (
	"net/http"

	"github.com/nimyab/nim2book-back/pkg/jwt"

	"github.com/labstack/echo/v4"
)

// HTTPv1 godoc
// @Summary	Translate book. Only VIP users can translate books
// @Tags	translate
// @Security BearerAuth
// @Accept	multipart/form-data
// @Produce	application/json
// @Param	file	formData	file	true	"Upload file"
// @Param	from	formData	string	true	"Source lang"
// @Param	to		formData	string	true	"Target lang"
// @Success	201		{object}	Output
// @Router	/translate/book [post]
func HTTPv1(c echo.Context) error {
	userPayload := jwt.GetUserPayload(c)

	bookFile, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "file is required",
		})
	}
	input := new(Input)
	if err = c.Bind(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "invalid input",
		})
	}
	if err = c.Validate(input); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": err.Error(),
		})
	}

	output, err := service.TranslateBook(input, bookFile, userPayload.Id)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, output)
}

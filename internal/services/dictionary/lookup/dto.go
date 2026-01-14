package lookup

import "github.com/nimyab/nim2book-back/internal/models"

type Input struct {
	Text string `json:"text" validate:"required"`
	Lang string `json:"lang" validate:"required"`
	UI   string `json:"ui" validate:"required"`
}

type Output = models.DictionaryData

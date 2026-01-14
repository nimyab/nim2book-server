package translate

import "github.com/nimyab/nim2book-back/internal/models"

type Input struct {
	Q            string               `json:"q" validate:"required"`
	Source       models.SupportedLang `json:"source" validate:"required"`
	Target       models.SupportedLang `json:"target" validate:"required"`
	Format       *string              `json:"format"`
	Alternatives *int                 `json:"alternatives"`
}

type Output struct {
	TranslatedText string `json:"translatedText"`
}

package translate

import "github.com/nimyab/nim2book-back/internal/domain"

type Input struct {
	Q            string               `json:"q" validate:"required"`
	Source       domain.SupportedLang `json:"source" validate:"required"`
	Target       domain.SupportedLang `json:"target" validate:"required"`
	Format       *string              `json:"format"`
	Alternatives *int                 `json:"alternatives"`
}

type Output struct {
	TranslatedText string `json:"translatedText"`
}

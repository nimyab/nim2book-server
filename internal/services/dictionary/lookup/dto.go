package lookup

import "github.com/nimyab/nim2book-back/internal/domain"

type Input struct {
	Text     string `json:"text" validate:"required"`
	FromLang string `json:"fromLang" validate:"required"`
	ToLang   string `json:"toLang" validate:"required"`
}

type Output struct {
	Words []*domain.DictionaryWord `json:"words"`
}

package lookup

// DictionaryData represents the response from Yandex Dictionary API
// This structure is kept for compatibility with external API
type DictionaryData struct {
	Head        map[string]any `json:"head,omitempty"`
	Definitions []Definition   `json:"def" validate:"required"`
}

type Definition struct {
	Text          string        `json:"text" validate:"required"`
	PartOfSpeech  string        `json:"pos"`
	Transcription string        `json:"ts"`
	Translations  []Translation `json:"tr" validate:"required"`
}

type Translation struct {
	Text         string    `json:"text" validate:"required"`
	PartOfSpeech string    `json:"pos"`
	Means        []Mean    `json:"mean"`
	Examples     []Example `json:"ex"`
}

type Mean struct {
	Text string `json:"text" validate:"required"`
}

type Example struct {
	Text        string               `json:"text" validate:"required"`
	Translation []ExampleTranslation `json:"tr"`
}

type ExampleTranslation struct {
	Text string `json:"text" validate:"required"`
}

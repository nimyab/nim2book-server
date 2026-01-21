package domain

import "github.com/google/uuid"

// DictionaryWord represents a word entry in the database
type DictionaryWord struct {
	ID            uuid.UUID           `json:"id"`
	Text          string              `json:"text"`
	FromLangCode  string              `json:"fromLangCode"`
	ToLangCode    string              `json:"toLangCode"`
	PartOfSpeech  string              `json:"partOfSpeech"`
	Translations  []string            `json:"translations"`
	Transcription *string             `json:"transcription"`
	Examples      []DictionaryExample `json:"examples"`
}

// DictionaryExample represents an example usage of a word
type DictionaryExample struct {
	ID                uuid.UUID `json:"id"`
	Text              string    `json:"text"`
	TranslatedText    string    `json:"translatedText"`
	WordPositionStart int       `json:"wordPositionStart"`
	WordPositionEnd   int       `json:"wordPositionEnd"`
	DictionaryID      uuid.UUID `json:"dictionaryID"`
}

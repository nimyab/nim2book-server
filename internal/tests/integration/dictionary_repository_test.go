package integration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDictionaryRepository_CreateDictionaryData(t *testing.T) {
	cleanupDatabase(t)

	content := []byte(`{"def":[{"text":"hello","tr":[{"text":"привет"}]}]}`)
	dict, err := dictRepo.CreateDictionaryData(context.Background(), "hello", "en-ru", content)
	require.NoError(t, err)
	assert.NotNil(t, dict)
	assert.NotEqual(t, uuid.Nil, dict.ID)
	assert.Equal(t, "hello", dict.Text)
	assert.Equal(t, "en-ru", dict.Lang)
	assert.Equal(t, content, dict.Content)
}

func TestDictionaryRepository_GetDictionaryData(t *testing.T) {
	cleanupDatabase(t)

	content := []byte(`{"def":[{"text":"world","tr":[{"text":"мир"}]}]}`)
	createTestDictionary(t, "world", "en-ru", content)

	dict, err := dictRepo.GetDictionaryData(context.Background(), "world", "en-ru")
	require.NoError(t, err)
	assert.Equal(t, "world", dict.Text)
	assert.Equal(t, "en-ru", dict.Lang)

	// Compare JSON content semantically instead of byte-by-byte
	var expected, actual map[string]interface{}
	require.NoError(t, json.Unmarshal(content, &expected))
	require.NoError(t, json.Unmarshal(dict.Content, &actual))
	assert.Equal(t, expected, actual)
}

func TestDictionaryRepository_GetDictionaryData_NotFound(t *testing.T) {
	cleanupDatabase(t)

	_, err := dictRepo.GetDictionaryData(context.Background(), "nonexistent", "en-ru")
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrDictionaryDataNotFound, err)
}

func TestDictionaryRepository_UpdateDictionaryData(t *testing.T) {
	cleanupDatabase(t)

	content1 := []byte(`{"def":[{"text":"update"}]}`)
	dict := createTestDictionary(t, "update", "en-ru", content1)

	content2 := []byte(`{"def":[{"text":"update","tr":[{"text":"обновить"}]}]}`)
	err := dictRepo.UpdateDictionaryData(context.Background(), dict.ID, content2)
	require.NoError(t, err)

	updated, err := dictRepo.GetDictionaryData(context.Background(), "update", "en-ru")
	require.NoError(t, err)

	// Compare JSON content semantically instead of byte-by-byte
	var expected, actual map[string]interface{}
	require.NoError(t, json.Unmarshal(content2, &expected))
	require.NoError(t, json.Unmarshal(updated.Content, &actual))
	assert.Equal(t, expected, actual)
}

func TestDictionaryRepository_DeleteDictionaryData(t *testing.T) {
	cleanupDatabase(t)

	content := []byte(`{"def":[{"text":"delete"}]}`)
	dict := createTestDictionary(t, "delete", "en-ru", content)

	err := dictRepo.DeleteDictionaryData(context.Background(), dict.ID)
	require.NoError(t, err)

	_, err = dictRepo.GetDictionaryData(context.Background(), "delete", "en-ru")
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrDictionaryDataNotFound, err)
}

func TestDictionaryRepository_ParseContent(t *testing.T) {
	cleanupDatabase(t)

	dictData := models.DictionaryData{
		Definitions: []models.Definition{
			{
				Text:         "test",
				PartOfSpeech: "noun",
				Translations: []models.Translation{
					{Text: "тест"},
				},
			},
		},
	}
	content, err := json.Marshal(dictData)
	require.NoError(t, err)

	dict := createTestDictionary(t, "test", "en-ru", content)

	parsed, err := dictRepo.ParseContent(dict)
	require.NoError(t, err)
	assert.Len(t, parsed.Definitions, 1)
	assert.Equal(t, "test", parsed.Definitions[0].Text)
	assert.Equal(t, "noun", parsed.Definitions[0].PartOfSpeech)
	assert.Len(t, parsed.Definitions[0].Translations, 1)
	assert.Equal(t, "тест", parsed.Definitions[0].Translations[0].Text)
}

func TestDictionaryRepository_MarshalContent(t *testing.T) {
	cleanupDatabase(t)

	dictData := &models.DictionaryData{
		Definitions: []models.Definition{
			{
				Text:         "marshal",
				PartOfSpeech: "verb",
				Translations: []models.Translation{
					{Text: "маршалировать"},
				},
			},
		},
	}

	content, err := dictRepo.MarshalContent(dictData)
	require.NoError(t, err)
	assert.NotEmpty(t, content)

	// Verify it can be unmarshaled back
	var parsed models.DictionaryData
	err = json.Unmarshal(content, &parsed)
	require.NoError(t, err)
	assert.Equal(t, "marshal", parsed.Definitions[0].Text)
}

func TestDictionaryRepository_FindDictionaryByLanguage(t *testing.T) {
	cleanupDatabase(t)

	createTestDictionary(t, "word1", "en-ru", []byte(`{}`))
	createTestDictionary(t, "word2", "en-ru", []byte(`{}`))
	createTestDictionary(t, "word3", "fr-ru", []byte(`{}`))

	dicts, total, err := dictRepo.FindDictionaryByLanguage(context.Background(), "en-ru", 1, 10)
	require.NoError(t, err)
	assert.Len(t, dicts, 2)
	assert.Equal(t, int64(2), total)
}

func TestDictionaryRepository_FindDictionaryByLanguage_Pagination(t *testing.T) {
	cleanupDatabase(t)

	// Create 5 entries
	for i := 1; i <= 5; i++ {
		createTestDictionary(t, "word"+string(rune('0'+i)), "en-ru", []byte(`{}`))
	}

	// Get first page (2 items)
	dicts, total, err := dictRepo.FindDictionaryByLanguage(context.Background(), "en-ru", 1, 2)
	require.NoError(t, err)
	assert.Len(t, dicts, 2)
	assert.Equal(t, int64(5), total)

	// Get second page
	dicts, total, err = dictRepo.FindDictionaryByLanguage(context.Background(), "en-ru", 2, 2)
	require.NoError(t, err)
	assert.Len(t, dicts, 2)
	assert.Equal(t, int64(5), total)

	// Get third page
	dicts, total, err = dictRepo.FindDictionaryByLanguage(context.Background(), "en-ru", 3, 2)
	require.NoError(t, err)
	assert.Len(t, dicts, 1)
	assert.Equal(t, int64(5), total)
}

func TestDictionaryRepository_SearchDictionary(t *testing.T) {
	cleanupDatabase(t)

	createTestDictionary(t, "apple", "en-ru", []byte(`{}`))
	createTestDictionary(t, "application", "en-ru", []byte(`{}`))
	createTestDictionary(t, "banana", "en-ru", []byte(`{}`))

	dicts, err := dictRepo.SearchDictionary(context.Background(), "app", "en-ru", 10)
	require.NoError(t, err)
	assert.Len(t, dicts, 2)
}

func TestDictionaryRepository_SearchDictionary_CaseInsensitive(t *testing.T) {
	cleanupDatabase(t)

	createTestDictionary(t, "Hello", "en-ru", []byte(`{}`))
	createTestDictionary(t, "WORLD", "en-ru", []byte(`{}`))

	dicts, err := dictRepo.SearchDictionary(context.Background(), "hello", "en-ru", 10)
	require.NoError(t, err)
	assert.Len(t, dicts, 1)
	assert.Equal(t, "Hello", dicts[0].Text)
}

func TestDictionaryRepository_CountDictionaryByLanguage(t *testing.T) {
	cleanupDatabase(t)

	createTestDictionary(t, "word1", "en-ru", []byte(`{}`))
	createTestDictionary(t, "word2", "en-ru", []byte(`{}`))
	createTestDictionary(t, "word3", "fr-ru", []byte(`{}`))

	count, err := dictRepo.CountDictionaryByLanguage(context.Background(), "en-ru")
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	count, err = dictRepo.CountDictionaryByLanguage(context.Background(), "fr-ru")
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestDictionaryRepository_GetSupportedLanguages(t *testing.T) {
	cleanupDatabase(t)

	createTestDictionary(t, "word1", "en-ru", []byte(`{}`))
	createTestDictionary(t, "word2", "fr-ru", []byte(`{}`))
	createTestDictionary(t, "word3", "de-ru", []byte(`{}`))
	createTestDictionary(t, "word4", "en-ru", []byte(`{}`)) // duplicate lang

	languages, err := dictRepo.GetSupportedLanguages(context.Background())
	require.NoError(t, err)
	assert.Len(t, languages, 3)
	assert.Contains(t, languages, "en-ru")
	assert.Contains(t, languages, "fr-ru")
	assert.Contains(t, languages, "de-ru")
}

func TestDictionaryRepository_DuplicateDictionary(t *testing.T) {
	cleanupDatabase(t)

	text := "duplicate"
	lang := "en-ru"
	content := []byte(`{}`)

	_, err := dictRepo.CreateDictionaryData(context.Background(), text, lang, content)
	require.NoError(t, err)

	_, err = dictRepo.CreateDictionaryData(context.Background(), text, lang, content)
	assert.Error(t, err)
	assert.Equal(t, repositories.ErrDictionaryDataAlreadyExists, err)
}

func TestDictionaryRepository_BulkCreateDictionary(t *testing.T) {
	cleanupDatabase(t)

	entries := []struct {
		Text    string
		Lang    string
		Content []byte
	}{
		{"word1", "en-ru", []byte(`{"def":[{"text":"word1"}]}`)},
		{"word2", "en-ru", []byte(`{"def":[{"text":"word2"}]}`)},
		{"word3", "en-ru", []byte(`{"def":[{"text":"word3"}]}`)},
	}

	created, err := dictRepo.BulkCreateDictionary(context.Background(), entries)
	require.NoError(t, err)
	assert.Equal(t, 3, created)

	// Verify all were created
	count, err := dictRepo.CountDictionaryByLanguage(context.Background(), "en-ru")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestDictionaryRepository_BulkCreate_SkipDuplicates(t *testing.T) {
	cleanupDatabase(t)

	// Pre-create one entry
	createTestDictionary(t, "word1", "en-ru", []byte(`{}`))

	entries := []struct {
		Text    string
		Lang    string
		Content []byte
	}{
		{"word1", "en-ru", []byte(`{}`)}, // duplicate - should skip
		{"word2", "en-ru", []byte(`{}`)}, // new
		{"word3", "en-ru", []byte(`{}`)}, // new
	}

	created, err := dictRepo.BulkCreateDictionary(context.Background(), entries)
	require.NoError(t, err)
	assert.Equal(t, 2, created) // Only 2 new entries created

	// Verify total count
	count, err := dictRepo.CountDictionaryByLanguage(context.Background(), "en-ru")
	require.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestDictionaryRepository_ComplexJSONContent(t *testing.T) {
	cleanupDatabase(t)

	complexData := models.DictionaryData{
		Head: map[string]interface{}{
			"lang": "en-ru",
		},
		Definitions: []models.Definition{
			{
				Text:          "complex",
				PartOfSpeech:  "adjective",
				Transcription: "ˈkɒmpleks",
				Translations: []models.Translation{
					{
						Text:         "сложный",
						PartOfSpeech: "прилагательное",
						Means: []models.Mean{
							{Text: "трудный"},
							{Text: "запутанный"},
						},
						Examples: []models.Example{
							{
								Text: "It's a complex problem",
								Translation: []models.ExampleTranslation{
									{Text: "Это сложная проблема"},
								},
							},
						},
					},
				},
			},
		},
	}

	content, err := json.Marshal(complexData)
	require.NoError(t, err)

	dict := createTestDictionary(t, "complex", "en-ru", content)

	parsed, err := dictRepo.ParseContent(dict)
	require.NoError(t, err)
	assert.Equal(t, "complex", parsed.Definitions[0].Text)
	assert.Equal(t, "ˈkɒmpleks", parsed.Definitions[0].Transcription)
	assert.Len(t, parsed.Definitions[0].Translations, 1)
	assert.Len(t, parsed.Definitions[0].Translations[0].Means, 2)
	assert.Len(t, parsed.Definitions[0].Translations[0].Examples, 1)
}

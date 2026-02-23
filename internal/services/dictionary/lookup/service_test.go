package lookup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLookup(t *testing.T) {
	tests := []struct {
		name           string
		inputText      string
		inputFromLang  string
		inputToLang    string
		mockDictRepo   func(*MockDictionaryRepo)
		mockRedis      func(*MockRedis)
		mockHTTPClient func(*MockHTTPClient)
		expectedOutput *Output
		expectedError  error
	}{
		{
			name:          "Cache Hit",
			inputText:     "hello",
			inputFromLang: "en",
			inputToLang:   "ru",
			mockDictRepo:  func(m *MockDictionaryRepo) {},
			mockRedis: func(m *MockRedis) {
				words := []*domain.DictionaryWord{{Text: "hello", Translations: []string{"привет"}}}
				data, _ := json.Marshal(words)
				m.On("Get", mock.Anything, "hello:en_ru").Return(data, nil)
			},
			mockHTTPClient: func(m *MockHTTPClient) {},
			expectedOutput: &Output{
				Words: []*domain.DictionaryWord{{Text: "hello", Translations: []string{"привет"}}},
			},
			expectedError: nil,
		},
		{
			name:          "Cache Miss, DB Hit",
			inputText:     "hello",
			inputFromLang: "en",
			inputToLang:   "ru",
			mockDictRepo: func(m *MockDictionaryRepo) {
				words := []*domain.DictionaryWord{{Text: "hello", Translations: []string{"привет"}}}
				m.On("GetDictionaryWordsByText", mock.Anything, "hello", "en", "ru").Return(words, nil)
			},
			mockRedis: func(m *MockRedis) {
				m.On("Get", mock.Anything, "hello:en_ru").Return(nil, errors.New("cache miss"))
				m.On("Save", mock.Anything, "hello:en_ru", mock.Anything, RedisCacheTTL).Return(nil)
			},
			mockHTTPClient: func(m *MockHTTPClient) {},
			expectedOutput: &Output{
				Words: []*domain.DictionaryWord{{Text: "hello", Translations: []string{"привет"}}},
			},
			expectedError: nil,
		},
		{
			name:          "Cache Miss, DB Miss, Yandex API Hit",
			inputText:     "hello",
			inputFromLang: "en",
			inputToLang:   "ru",
			mockDictRepo: func(m *MockDictionaryRepo) {
				m.On("GetDictionaryWordsByText", mock.Anything, "hello", "en", "ru").Return(nil, nil)
				expectedWord := &domain.DictionaryWord{
					Text:         "hello",
					Translations: []string{"привет"},
					PartOfSpeech: "noun",
				}
				m.On("Create", mock.Anything, mock.MatchedBy(func(w *domain.DictionaryWord) bool {
					return w.Text == "hello"
				})).Return(expectedWord, nil)
			},
			mockRedis: func(m *MockRedis) {
				m.On("Get", mock.Anything, "hello:en_ru").Return(nil, errors.New("cache miss"))
				m.On("Save", mock.Anything, "hello:en_ru", mock.Anything, RedisCacheTTL).Return(nil)
			},
			mockHTTPClient: func(m *MockHTTPClient) {
				responseBody := `{
					"def": [
						{
							"text": "hello",
							"pos": "noun",
							"tr": [
								{"text": "привет", "pos": "noun"}
							]
						}
					]
				}`
				resp := &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(responseBody)),
				}
				m.On("Do", mock.Anything).Return(resp, nil)
			},
			expectedOutput: &Output{
				Words: []*domain.DictionaryWord{{Text: "hello", Translations: []string{"привет"}, PartOfSpeech: "noun"}},
			},
			expectedError: nil,
		},
		{
			name:          "Cache Miss, DB Miss, Yandex API Error",
			inputText:     "hello",
			inputFromLang: "en",
			inputToLang:   "ru",
			mockDictRepo: func(m *MockDictionaryRepo) {
				m.On("GetDictionaryWordsByText", mock.Anything, "hello", "en", "ru").Return(nil, nil)
			},
			mockRedis: func(m *MockRedis) {
				m.On("Get", mock.Anything, "hello:en_ru").Return(nil, errors.New("cache miss"))
			},
			mockHTTPClient: func(m *MockHTTPClient) {
				m.On("Do", mock.Anything).Return(nil, errors.New("api error"))
			},
			expectedOutput: nil,
			expectedError:  ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDictRepo := NewMockDictionaryRepo(t)
			mockRedis := NewMockRedis(t)
			mockHTTPClient := NewMockHTTPClient(t)

			if tt.mockDictRepo != nil {
				tt.mockDictRepo(mockDictRepo)
			}
			if tt.mockRedis != nil {
				tt.mockRedis(mockRedis)
			}
			if tt.mockHTTPClient != nil {
				tt.mockHTTPClient(mockHTTPClient)
			}

			service := New(mockDictRepo, mockRedis, mockHTTPClient, "key", "url")

			input := &Input{
				Text:     tt.inputText,
				FromLang: tt.inputFromLang,
				ToLang:   tt.inputToLang,
			}

			output, err := service.Lookup(context.Background(), input)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, output)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, output)
				assert.Equal(t, len(tt.expectedOutput.Words), len(output.Words))
				if len(output.Words) > 0 {
					assert.Equal(t, tt.expectedOutput.Words[0].Text, output.Words[0].Text)
					assert.Equal(t, tt.expectedOutput.Words[0].Translations, output.Words[0].Translations)
				}
			}
		})
	}
}

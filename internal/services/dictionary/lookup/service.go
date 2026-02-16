package lookup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/google/uuid"
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/logger"
)

type Postgres interface {
	CreateDictionaryWord(context.Context, *domain.DictionaryWord) (domain.ID, error)
	CreateDictionaryExample(context.Context, *domain.DictionaryExample) (domain.ID, error)
	GetDictionaryWordsByText(ctx context.Context, text, fromLang, toLang string) ([]*domain.DictionaryWord, error)
}

type Redis interface {
	Save(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
}

type Service struct {
	pg            Postgres
	redis         Redis
	yandexDictKey string
	yandexDictURL string
}

func New(pg Postgres, redis Redis, yandexDictKey, yandexDictURL string) *Service {
	return &Service{
		pg:            pg,
		redis:         redis,
		yandexDictKey: yandexDictKey,
		yandexDictURL: yandexDictURL,
	}
}

var (
	ErrInternal = errors.New("internal server error")
)

const (
	RedisCacheTTL = 15 * time.Minute
)

func (s *Service) Lookup(input *Input) (*Output, error) {
	const operation = "lookup.Lookup"
	ctx := context.Background()

	// Нормализуем входной текст
	normalizedText := strings.ToLower(strings.Trim(input.Text, ".!?,;: "))
	redisKey := fmt.Sprintf("%s:%s_%s", normalizedText, input.FromLang, input.ToLang)

	// Проверяем кеш
	if words, ok := s.getFromCache(ctx, redisKey); ok {
		return &Output{Words: words}, nil
	}

	// Проверяем базу данных
	if words, ok := s.getFromDatabase(ctx, normalizedText, input.FromLang, input.ToLang, redisKey); ok {
		return &Output{Words: words}, nil
	}

	// Получаем данные из внешнего API
	words, err := s.fetchAndSaveFromYandex(ctx, normalizedText, input.FromLang, input.ToLang, redisKey)
	if err != nil {
		logger.Error("failed to fetch from yandex", err, operation)
		return nil, ErrInternal
	}

	return &Output{Words: words}, nil
}

// getFromCache пытается получить данные из кеша
func (s *Service) getFromCache(ctx context.Context, redisKey string) ([]domain.DictionaryWord, bool) {
	const operation = "lookup.getFromCache"

	byteData, err := s.redis.Get(ctx, redisKey)
	if err != nil {
		slog.Info("failed to get from redis", slog.String("redisKey", redisKey), slog.Any("error", err))
		return nil, false
	}

	var dictWords []domain.DictionaryWord
	if err = json.Unmarshal(byteData, &dictWords); err != nil {
		logger.Error("failed to unmarshal json", err, operation)
		return nil, false
	}

	if len(dictWords) == 0 {
		slog.Info("dictionary cache empty", slog.String("redisKey", redisKey))
		return nil, false
	}

	slog.Info("dictionary cache hit", slog.String("redisKey", redisKey))
	return dictWords, true
}

// getFromDatabase пытается получить данные из базы данных
func (s *Service) getFromDatabase(ctx context.Context, text, fromLang, toLang, redisKey string) ([]domain.DictionaryWord, bool) {
	slog.Info("dictionary cache miss, checking database", slog.String("redisKey", redisKey))

	dictDataPtrs, err := s.pg.GetDictionaryWordsByText(ctx, text, fromLang, toLang)
	if err != nil {
		slog.Info("failed to get from postgres", slog.Any("error", err))
		return nil, false
	}

	if len(dictDataPtrs) == 0 {
		return nil, false
	}

	// Конвертируем []*DictionaryWord в []DictionaryWord
	dictData := make([]domain.DictionaryWord, len(dictDataPtrs))
	for i, ptr := range dictDataPtrs {
		dictData[i] = *ptr
	}

	slog.Info("dictionary database hit", slog.String("redisKey", redisKey))
	s.saveToCache(ctx, dictData, redisKey)
	return dictData, true
}

// fetchAndSaveFromYandex получает данные из Yandex API и сохраняет их
func (s *Service) fetchAndSaveFromYandex(ctx context.Context, text, fromLang, toLang, redisKey string) ([]domain.DictionaryWord, error) {
	const operation = "lookup.fetchAndSaveFromYandex"
	slog.Info("dictionary database miss, fetching from yandex api", slog.String("redisKey", redisKey))

	// Делаем запрос к Yandex API
	yandexResponse, err := s.fetchFromYandexAPI(text, fromLang, toLang)
	if err != nil {
		logger.Error("failed to fetch from yandex api", err, operation)
		return nil, err
	}

	slog.Info("yandex dictionary response", slog.Any("definitions_count", len(yandexResponse.Definitions)))

	// Конвертируем и сохраняем данные
	words, err := s.convertAndSaveYandexResponse(ctx, yandexResponse, fromLang, toLang)
	if err != nil {
		logger.Error("failed to convert and save yandex response", err, operation)
		return nil, err
	}

	// Сохраняем в кеш
	s.saveToCache(ctx, words, redisKey)

	return words, nil
}

// fetchFromYandexAPI делает запрос к Yandex Dictionary API
func (s *Service) fetchFromYandexAPI(text, fromLang, toLang string) (*DictionaryData, error) {
	const operation = "lookup.fetchFromYandexAPI"

	resp, err := retry.DoWithData(func() (*http.Response, error) {
		baseURL := fmt.Sprintf("%s/%s", s.yandexDictURL, "lookup")
		req, err := http.NewRequest("GET", baseURL, nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Add("key", s.yandexDictKey)
		q.Add("lang", fmt.Sprintf("%s-%s", fromLang, toLang))
		q.Add("text", text)
		req.URL.RawQuery = q.Encode()

		return http.DefaultClient.Do(req)
	}, retry.Attempts(5))

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var yandexResponse DictionaryData
	if err = json.Unmarshal(body, &yandexResponse); err != nil {
		logger.Error("failed to unmarshal response body", err, operation)
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &yandexResponse, nil
}

// convertAndSaveYandexResponse конвертирует ответ Yandex в доменные объекты и сохраняет в БД
func (s *Service) convertAndSaveYandexResponse(ctx context.Context, yandexResponse *DictionaryData, fromLang, toLang string) ([]domain.DictionaryWord, error) {
	const operation = "lookup.convertAndSaveYandexResponse"
	var words []domain.DictionaryWord

	for _, definition := range yandexResponse.Definitions {
		if !s.isSupportedPartOfSpeech(definition.PartOfSpeech) {
			slog.Info(fmt.Sprintf("skip %s word, because part of speech is %s", definition.Text, definition.PartOfSpeech))
			continue
		}

		word, err := s.saveWordWithExamples(ctx, definition, fromLang, toLang)
		if err != nil {
			logger.Error("failed to save word with examples", err, operation)
			continue
		}
		words = append(words, word)
	}

	return words, nil
}

// saveWordWithExamples сохраняет слово и его примеры в базу данных
func (s *Service) saveWordWithExamples(ctx context.Context, definition Definition, fromLang, toLang string) (domain.DictionaryWord, error) {
	const operation = "lookup.saveWordWithExamples"

	// Собираем все переводы
	wordTranslations := make([]string, 0, len(definition.Translations))
	for _, translation := range definition.Translations {
		wordTranslations = append(wordTranslations, translation.Text)
	}

	// Создаем объект слова
	wordData := domain.DictionaryWord{
		Text:          definition.Text,
		FromLangCode:  fromLang,
		ToLangCode:    toLang,
		PartOfSpeech:  definition.PartOfSpeech,
		Transcription: &definition.Transcription,
		Translations:  wordTranslations,
	}

	// Сохраняем слово
	wordID, err := s.pg.CreateDictionaryWord(ctx, &wordData)
	if err != nil {
		return domain.DictionaryWord{}, fmt.Errorf("failed to create word: %w", err)
	}
	wordData.ID = wordID

	// Сохраняем примеры
	examples, err := s.saveExamples(ctx, definition.Translations, wordID)
	if err != nil {
		logger.Error("failed to save examples", err, operation)
	}
	wordData.Examples = examples

	return wordData, nil
}

// saveExamples сохраняет примеры использования слова
func (s *Service) saveExamples(ctx context.Context, translations []Translation, wordID uuid.UUID) ([]domain.DictionaryExample, error) {
	const operation = "lookup.saveExamples"
	var examples []domain.DictionaryExample

	for _, translation := range translations {
		for _, example := range translation.Examples {
			if len(example.Translation) == 0 {
				continue
			}

			exampleData := domain.DictionaryExample{
				Text:           example.Text,
				TranslatedText: example.Translation[0].Text,
				DictionaryID:   wordID,
			}

			exampleID, err := s.pg.CreateDictionaryExample(ctx, &exampleData)
			if err != nil {
				logger.Error("failed to create dictionary example", err, operation)
				continue
			}
			exampleData.ID = exampleID

			examples = append(examples, exampleData)
		}
	}

	return examples, nil
}

// saveToCache сохраняет данные в кеш
func (s *Service) saveToCache(ctx context.Context, words []domain.DictionaryWord, redisKey string) {
	const operation = "lookup.saveToCache"

	if len(words) == 0 {
		return
	}

	byteData, err := json.Marshal(words)
	if err != nil {
		logger.Error("failed to marshal json", err, operation)
		return
	}

	if err = s.redis.Save(ctx, redisKey, byteData, RedisCacheTTL); err != nil {
		logger.Error("failed to save to redis", err, operation)
	}
}

func (s *Service) isSupportedPartOfSpeech(partOfSpeech string) bool {
	switch partOfSpeech {
	case "interjection":
	case "noun":
	case "verb":
	case "pronoun":
	case "preposition":
	case "adverb":
	case "participle":
	case "adjective":
		return true
	}
	return false
}

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
	"github.com/nimyab/nim2book-back/internal/domain"
	"github.com/nimyab/nim2book-back/pkg/logger"
)

type DictionaryRepo interface {
	Create(ctx context.Context, domainDict *domain.DictionaryWord) (*domain.DictionaryWord, error)
	GetDictionaryWordsByText(ctx context.Context, text, fromLang, toLang string) ([]*domain.DictionaryWord, error)
}

type Redis interface {
	Save(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
}

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Service struct {
	dictRepo      DictionaryRepo
	redis         Redis
	httpClient    HTTPClient
	yandexDictKey string
	yandexDictURL string
}

func New(dictRepo DictionaryRepo, redis Redis, httpClient HTTPClient, yandexDictKey, yandexDictURL string) *Service {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &Service{
		dictRepo:      dictRepo,
		redis:         redis,
		httpClient:    httpClient,
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

func (s *Service) Lookup(ctx context.Context, input *Input) (*Output, error) {
	const operation = "lookup.Lookup"

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
func (s *Service) getFromCache(ctx context.Context, redisKey string) ([]*domain.DictionaryWord, bool) {
	const operation = "lookup.getFromCache"

	byteData, err := s.redis.Get(ctx, redisKey)
	if err != nil {
		slog.Info("failed to get from redis", slog.String("redisKey", redisKey), slog.Any("error", err))
		return nil, false
	}

	var dictWords []*domain.DictionaryWord
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
func (s *Service) getFromDatabase(ctx context.Context, text, fromLang, toLang, redisKey string) ([]*domain.DictionaryWord, bool) {
	slog.Info("dictionary cache miss, checking database", slog.String("redisKey", redisKey))

	dictData, err := s.dictRepo.GetDictionaryWordsByText(ctx, text, fromLang, toLang)
	if err != nil {
		slog.Info("failed to get from postgres", slog.Any("error", err))
		return nil, false
	}

	if len(dictData) == 0 {
		return nil, false
	}

	slog.Info("dictionary database hit", slog.String("redisKey", redisKey))
	s.saveToCache(ctx, dictData, redisKey)
	return dictData, true
}

// fetchAndSaveFromYandex получает данные из Yandex API и сохраняет их
func (s *Service) fetchAndSaveFromYandex(ctx context.Context, text, fromLang, toLang, redisKey string) ([]*domain.DictionaryWord, error) {
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

		return s.httpClient.Do(req)
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
func (s *Service) convertAndSaveYandexResponse(ctx context.Context, yandexResponse *DictionaryData, fromLang, toLang string) ([]*domain.DictionaryWord, error) {
	const operation = "lookup.convertAndSaveYandexResponse"
	var words []*domain.DictionaryWord

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
		words = append(words, &word)
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

	// Сохраняем примеры и получаем их доменные объекты
	var examples []domain.DictionaryExample
	for _, translation := range definition.Translations {
		for _, example := range translation.Examples {
			if len(example.Translation) == 0 {
				continue
			}
			examples = append(examples, domain.DictionaryExample{
				Text:           example.Text,
				TranslatedText: example.Translation[0].Text,
			})
		}
	}

	// Создаем объект слова
	wordData := domain.DictionaryWord{
		Text:          definition.Text,
		FromLangCode:  fromLang,
		ToLangCode:    toLang,
		PartOfSpeech:  definition.PartOfSpeech,
		Transcription: &definition.Transcription,
		Translations:  wordTranslations,
		Examples:      examples,
	}

	// Сохраняем слово
	word, err := s.dictRepo.Create(ctx, &wordData)
	if err != nil {
		return domain.DictionaryWord{}, fmt.Errorf("failed to create word: %w", err)
	}

	return *word, nil
}

// saveToCache сохраняет данные в кеш
func (s *Service) saveToCache(ctx context.Context, words []*domain.DictionaryWord, redisKey string) {
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
	case "interjection", "noun", "verb", "pronoun", "preposition", "adverb", "participle", "adjective":
		return true
	}
	return false
}

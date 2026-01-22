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
	CreateDictionaryWord(context.Context, *domain.DictionaryWord) (uuid.UUID, error)
	CreateDictionaryExample(context.Context, *domain.DictionaryExample) (uuid.UUID, error)
	GetDictionaryWordsByText(ctx context.Context, text, fromLang, toLang string) ([]domain.DictionaryWord, error)
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

var service *Service

func New(pg Postgres, redis Redis, yandexDictKey, yandexDictURL string) *Service {
	service = &Service{
		pg:            pg,
		redis:         redis,
		yandexDictKey: yandexDictKey,
		yandexDictURL: yandexDictURL,
	}
	return service
}

var (
	ErrInternal = errors.New("internal server error")
)

const (
	RedisCacheTTL = 15 * time.Minute
)

func (s *Service) Lookup(input *Input) (*Output, error) {
	const operation = "lookup.Lookup"

	input.Text = strings.ToLower(strings.Trim(input.Text, ".!?,;: "))
	var redisKey = fmt.Sprintf("%s:%s_%s", input.Text, input.FromLang, input.ToLang)

	byteData, err := s.redis.Get(context.Background(), redisKey)
	if err != nil {
		logger.Error("failed to get from redis", err, operation)
	} else {
		var dictWords []domain.DictionaryWord
		err = json.Unmarshal(byteData, &dictWords)
		slog.Info("dictionary cache hit", slog.String("redisKey", redisKey), slog.Any("dictWords", dictWords))
		if err != nil {
			logger.Error("failed to unmarshal json", err, operation)
		} else if len(dictWords) == 0 {
			slog.Info("dictionary cache empty", slog.String("redisKey", redisKey))
		} else {
			return &Output{Words: dictWords}, nil
		}
	}

	slog.Info("dictionary cache miss, start get from database", slog.String("redisKey", redisKey))
	dictData, err := s.pg.GetDictionaryWordsByText(context.Background(), input.Text, input.FromLang, input.ToLang)
	if err != nil {
		logger.Error("failed to get from postgres", err, operation)
	}
	if len(dictData) > 0 {
		slog.Info("dictionary database hit", slog.String("redisKey", redisKey), slog.Any("dictData", dictData))
		s.saveToRedis(dictData, redisKey)
		return &Output{Words: dictData}, nil
	}

	slog.Info("dictionary database miss, start get from yandex api", slog.String("redisKey", redisKey))
	resp, err := retry.DoWithData(func() (*http.Response, error) {
		baseURL := fmt.Sprintf("%s/%s", s.yandexDictURL, "lookup")
		req, err := http.NewRequest("GET", baseURL, nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Add("key", s.yandexDictKey)
		q.Add("lang", fmt.Sprintf("%s-%s", input.FromLang, input.ToLang))
		q.Add("text", input.Text)
		req.URL.RawQuery = q.Encode()

		return http.DefaultClient.Do(req)
	}, retry.Attempts(5))

	if err != nil {
		logger.Error("failed to request to yandex dict", err, operation)
		return nil, ErrInternal
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error("failed to read response body", err, operation)
		return nil, ErrInternal
	}
	defer resp.Body.Close()

	yandexLookupResponse := DictionaryData{}
	if err = json.Unmarshal(body, &yandexLookupResponse); err != nil {
		logger.Error("failed to unmarshal response body", err, operation)
		return nil, ErrInternal
	}
	slog.Info("yandex dictionary len definitions", slog.Any("len definitions", len(yandexLookupResponse.Definitions)))

	var newDictData []domain.DictionaryWord

	for _, definition := range yandexLookupResponse.Definitions {
		wordTranslations := make([]string, 0)
		for _, translation := range definition.Translations {
			wordTranslations = append(wordTranslations, translation.Text)
		}

		wordData := domain.DictionaryWord{
			Text:          definition.Text,
			FromLangCode:  input.FromLang,
			ToLangCode:    input.ToLang,
			PartOfSpeech:  definition.PartOfSpeech,
			Transcription: &definition.Transcription,
			Translations:  wordTranslations,
		}

		wordId, err := s.pg.CreateDictionaryWord(context.Background(), &wordData)
		if err != nil {
			logger.Error("failed to create dictionary word", err, operation)
			continue
		}
		wordData.ID = wordId

		examplesData := make([]domain.DictionaryExample, 0)
		for _, translation := range definition.Translations {
			for _, example := range translation.Examples {
				exampleData := domain.DictionaryExample{
					Text:           example.Text,
					TranslatedText: example.Translation[0].Text,
					DictionaryID:   wordId,
				}

				exampleId, err := s.pg.CreateDictionaryExample(context.Background(), &exampleData)
				if err != nil {
					logger.Error("failed to create dictionary example", err, operation)
					continue
				}
				exampleData.ID = exampleId

				examplesData = append(examplesData, exampleData)
			}
		}

		wordData.Examples = examplesData

		newDictData = append(newDictData, wordData)
	}

	s.saveToRedis(newDictData, redisKey)

	return &Output{Words: newDictData}, nil
}

func (s *Service) saveToRedis(dictData []domain.DictionaryWord, redisKey string) {
	const operation = "lookup.saveToRedis"

	if len(dictData) == 0 {
		return
	}
	byteData, err := json.Marshal(dictData)
	if err != nil {
		logger.Error("failed to marshal json", err, operation)
	} else {
		if err = s.redis.Save(context.Background(), redisKey, byteData, RedisCacheTTL); err != nil {
			logger.Error("failed to save dictionary data to redis", err, operation)
		}
	}
}

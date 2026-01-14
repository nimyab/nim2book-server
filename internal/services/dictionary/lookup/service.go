package lookup

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/avast/retry-go/v4"
	"github.com/nimyab/nim2book-back/internal/models"
	"github.com/nimyab/nim2book-back/internal/repositories"
	"github.com/nimyab/nim2book-back/pkg/logger"
)

type Redis interface {
	Save(ctx context.Context, key string, value []byte, expiration time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
}

type Service struct {
	dictRepo      *repositories.DictionaryRepository
	redis         Redis
	yandexDictKey string
	yandexDictURL string
}

var service *Service

func New(dictRepo *repositories.DictionaryRepository, redis Redis, yandexDictKey, yandexDictURL string) *Service {
	service = &Service{
		dictRepo:      dictRepo,
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

	ctx := context.Background()
	output := new(models.DictionaryData)

	// Try to get from cache first
	byteData, err := s.redis.Get(ctx, fmt.Sprintf("%s:%s", input.Text, input.Lang))
	if err == nil {
		err = json.Unmarshal(byteData, output)
		if err == nil {
			return output, nil
		}
		logger.Error("failed to unmarshal from redis", err, operation)
	}

	// Try to get from database
	dict, err := s.dictRepo.GetDictionaryData(ctx, input.Text, input.Lang)
	if err == nil && dict != nil {
		// Parse dictionary content
		dictData, err := s.dictRepo.ParseContent(dict)
		if err != nil {
			logger.Error("failed to parse dictionary content", err, operation)
		} else {
			// Save to cache
			byteData, err = json.Marshal(dictData)
			if err != nil {
				logger.Error("failed to marshal json", err, operation)
			} else {
				if err = s.redis.Save(context.Background(), fmt.Sprintf("%s:%s", input.Text, input.Lang), byteData, RedisCacheTTL); err != nil {
					logger.Error("failed to save dictionary data to redis", err, operation)
				}
			}
			return dictData, nil
		}
	}

	resp, err := retry.DoWithData(func() (*http.Response, error) {
		baseURL := fmt.Sprintf("%s/%s", s.yandexDictURL, "lookup")
		req, err := http.NewRequest("GET", baseURL, nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Add("key", s.yandexDictKey)
		q.Add("lang", input.Lang)
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

	if err = json.Unmarshal(body, output); err != nil {
		logger.Error("failed to unmarshal response body", err, operation)
		return nil, ErrInternal
	}

	// Save to database
	content, err := json.Marshal(output)
	if err != nil {
		logger.Error("failed to marshal dictionary data", err, operation)
	} else {
		if _, err = s.dictRepo.CreateDictionaryData(context.Background(), input.Text, input.Lang, content); err != nil {
			logger.Error("failed to save dictionary data to postgres", err, operation)
		}
	}
	if err = s.redis.Save(context.Background(), fmt.Sprintf("%s:%s", input.Text, input.Lang), body, RedisCacheTTL); err != nil {
		logger.Error("failed to save dictionary data to redis", err, operation)
	}

	return output, nil
}

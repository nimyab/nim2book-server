package redis_cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	RedisURL string
}

type RedisCache struct {
	client *redis.Client
}

func New(cfg *Config) (*RedisCache, error) {
	const operation = "redis_cache.New"

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	client := redis.NewClient(opt)

	if _, err = client.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &RedisCache{client: client}, nil
}

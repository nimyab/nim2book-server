package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	RedisURL string
}

type Redis struct {
	client *redis.Client
}

func New(cfg *Config) (*Redis, error) {
	const operation = "redis.New"

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	client := redis.NewClient(opt)

	if _, err = client.Ping(context.Background()).Result(); err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &Redis{client: client}, nil
}

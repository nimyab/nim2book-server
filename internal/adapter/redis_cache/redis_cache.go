package redis_cache

import (
	"context"
	"fmt"
	"time"

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

func (r *RedisCache) Save(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	const operation = "redis_cache.Set"

	if err := r.client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}

	return nil
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	const operation = "redis_cache.Get"

	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return []byte(result), nil
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	const operation = "redis_cache.Delete"

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("%s: %w", operation, err)
	}
	
	return nil
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	const operation = "redis_cache.Exists"

	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("%s: %w", operation, err)
	}

	return count > 0, nil
}

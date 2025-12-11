package redis

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client *redis.Client
	logger domain.Logger
}

func NewRedisStorage(client *redis.Client, logger domain.Logger) *RedisStorage {
	return &RedisStorage{
		client: client,
		logger: logger,
	}
}

func (s *RedisStorage) GetOriginalUrl(ctx context.Context, shortUrl string) (string, bool) {
	val, err := s.client.Get(ctx, shortUrl).Result()
	if err == redis.Nil {
		return "", false
	} else if err != nil {
		s.logger.Error(fmt.Sprintf("Failed to get original URL from Redis: %v", err))
		return "", false
	}

	return val, true
}

func (s *RedisStorage) SetMapping(ctx context.Context, originalUrl, urlToken string) error {
	return s.client.Set(ctx, urlToken, originalUrl, 0).Err()
}

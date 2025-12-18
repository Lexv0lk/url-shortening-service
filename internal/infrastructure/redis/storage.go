package redis

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

// RedisStorage implements URL mapping cache operations using Redis.
// It provides fast read access to URL mappings with optional TTL support.
type RedisStorage struct {
	client domain.KeyStorage
	logger domain.Logger
}

// NewRedisStorage creates a new RedisStorage instance.
// Parameters:
//   - client: Redis client connection
//   - logger: logger for recording errors
func NewRedisStorage(client domain.KeyStorage, logger domain.Logger) *RedisStorage {
	return &RedisStorage{
		client: client,
		logger: logger,
	}
}

// GetOriginalUrl retrieves the original URL for a given short URL token from Redis.
// Returns the original URL and true if found, or empty string and false if not found.
// Redis errors are logged and result in returning false.
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

// SetMapping creates a new mapping between a URL token and its original URL in Redis.
// The mapping is stored without expiration (TTL = 0).
//
// Returns an error if the Redis SET operation fails.
func (s *RedisStorage) SetMapping(ctx context.Context, originalUrl, urlToken string) error {
	return s.client.Set(ctx, urlToken, originalUrl, 0).Err()
}

// DeleteMapping removes a URL mapping from Redis by its token.
//
// Returns an error if:
//   - *domain.TokenNonExistingError: the token does not exist in Redis
//   - Redis DEL operation fails
func (s *RedisStorage) DeleteMapping(ctx context.Context, urlToken string) error {
	n, err := s.client.Del(ctx, urlToken).Result()

	if err != nil {
		return err
	} else if n == 0 {
		return &domain.TokenNonExistingError{Msg: fmt.Sprintf("URL token not found in redis: %s", urlToken)}
	}

	return nil
}

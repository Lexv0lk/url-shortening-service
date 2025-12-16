package redis

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

const counterId = "mapping_count"

// RedisIdGenerator generates unique IDs for URL mappings using Redis INCR.
// It maintains an atomic counter in Redis to ensure uniqueness across instances.
type RedisIdGenerator struct {
	client *redis.Client
}

// NewRedisIdGenerator creates a new RedisIdGenerator instance.
// It initializes the Redis counter with the last known ID from persistent storage.
// Parameters:
//   - ctx: context for Redis operations
//   - client: Redis client connection
//   - lastIdGetter: retrieves the last used ID from persistent storage
//
// Returns an error if:
//   - Retrieving the last ID from storage fails
//   - Setting the initial counter value in Redis fails
func NewRedisIdGenerator(ctx context.Context, client *redis.Client, lastIdGetter domain.MappingInfoLastIdGetter) (*RedisIdGenerator, error) {
	lastId, err := lastIdGetter.GetLastId(ctx)
	if err != nil {
		return nil, fmt.Errorf("getting last mapping id: %w", err)
	}

	err = client.Set(ctx, counterId, lastId, 0).Err()
	if err != nil {
		return nil, fmt.Errorf("setting mapping count in redis: %w", err)
	}

	return &RedisIdGenerator{client: client}, nil
}

// GetNextId generates and returns the next unique ID for URL mappings.
// It atomically increments the Redis counter and returns the new value.
//
// Returns an error if the Redis INCR operation fails.
func (r *RedisIdGenerator) GetNextId(ctx context.Context) (int64, error) {
	newId, err := r.client.Incr(ctx, counterId).Result()
	if err != nil {
		return 0, fmt.Errorf("incrementing mapping count in redis: %w", err)
	}

	return newId, nil
}

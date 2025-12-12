package redis

import (
	"context"
	"fmt"
	"url-shortening-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

const counterId = "mapping_count"

type RedisIdGenerator struct {
	client *redis.Client
}

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

func (r *RedisIdGenerator) GetNextId(ctx context.Context) (int64, error) {
	newId, err := r.client.Incr(ctx, counterId).Result()
	if err != nil {
		return 0, fmt.Errorf("incrementing mapping count in redis: %w", err)
	}

	return newId, nil
}

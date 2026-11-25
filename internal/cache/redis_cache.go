package cache

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	Client *redis.Client
}

func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{
		Client: client,
	}
}

func (r *RedisCache) Get(
	ctx context.Context,
	key string,
	dest any,
) (bool, error) {
	value, err := r.Client.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	if err := json.Unmarshal([]byte(value), dest); err != nil {
		r.Client.Del(ctx, key)
		return false, nil
	}

	return true, nil
}

func (r *RedisCache) Set (
	ctx context.Context,
	key string,
	value any,
	ttl time.Duration,
 ) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return r.Client.Set(ctx, key, data, ttl).Err()
}

func (r *RedisCache) Delete(
	ctx context.Context,
	key string,
) error {
	return r.Client.Del(ctx, key).Err()
}
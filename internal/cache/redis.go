package cache

import (
	"github.com/go-redis/redis/v7"
	"github.com/pkg/errors"
	"time"
)

type Cacher interface {
	Has(key string) (bool, error)
	CreateKey(key string, ttl time.Duration) error
}

type RedisCache struct {
	store  *redis.Client
}

func NewRedisCache(store  *redis.Client) *RedisCache {
	return &RedisCache{
		store: store,
	}
}

func (c *RedisCache) Has(key string) (bool, error) {
	found, err := c.store.Exists(key).Result()
	if err != nil {
		return false, errors.Wrapf(err, "could not check existence for key %s", key)
	}

	return found == 1, nil
}

func (c *RedisCache) CreateKey(key string, ttl time.Duration) error {
	_, err := c.store.Set(key, 1, ttl).Result()
	if err != nil {
		return errors.Wrapf(err, "could not create key %s", key)
	}

	return nil
}



package cache

import (
	"encoding/json"
	"github.com/denismitr/auditbase/utils/errtype"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/go-redis/redis/v7"
	"github.com/pkg/errors"
	"time"
)

const ErrCouldNotPutValueToCache = errtype.StringError("could not put value to cache")
const ErrCouldNotRawValueToTarget = errtype.StringError("could not parse raw value and convert to target")
const ErrCouldNotCheckKeyExistence = errtype.StringError("could not check key existence")

type ResultFunc func() (interface{}, error)
type TargetParser func(v, target interface{}) error
type RememberFunc func(string, time.Duration, interface{}, ResultFunc) error

type Cacher interface {
	Remember(TargetParser) RememberFunc
	Has(key string) (bool, error)
	CreateKey(key string, ttl time.Duration) error
}

type RedisCache struct {
	store  *redis.Client
	log logger.Logger
}

func NewRedisCache(store  *redis.Client, log logger.Logger) *RedisCache {
	return &RedisCache{
		store: store,
		log: log,
	}
}

func (c *RedisCache) Remember(parser TargetParser) RememberFunc {
	return func(key string, ttl time.Duration, target interface{}, f ResultFunc) error {
		str, err := c.store.Get(key).Result()
		if str != "" {
			if err := json.Unmarshal([]byte(str), &target); err != nil {
				c.log.Error(errors.Wrapf(err, "could parse payload from cache with key [%s]", key))
			} else {
				return nil
			}
		}

		c.log.Debugf("cache miss for key %s", key)

		v, err := f()
		if err != nil {
			return err
		}

		if err := parser(v, target); err != nil {
			return ErrCouldNotRawValueToTarget
		}

		b, err := json.Marshal(target)
		if err != nil {
			c.log.Error(errors.Wrapf(err, "could not create payload from value with key [%s] to put to cache", key))
			return ErrCouldNotPutValueToCache
		}

		if _, err := c.store.Set(key, string(b), ttl).Result(); err != nil {
			return ErrCouldNotPutValueToCache
		}

		return nil
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



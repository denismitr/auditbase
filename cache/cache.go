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
const ErrCouldNotRawValueToTarget = errtype.StringError("could bot parse raw value and convert to target")

type ResultFunc func() (interface{}, error)
type TargetParser func(v, target interface{}) error
type RememberFunc func(string, time.Duration, interface{}, ResultFunc) error

type Cacher interface {
	RememberFunc(TargetParser) RememberFunc
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

func (c *RedisCache) RememberFunc(parser TargetParser) RememberFunc {
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



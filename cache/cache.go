package cache

import (
	"encoding/json"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/errtype"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/go-redis/redis/v7"
	"github.com/pkg/errors"
	"time"
)

const ErrCouldNotPutValueToCache = errtype.StringError("could not put value to cache")

type ResultFunc func() (interface{}, error)

type Cacher interface {
	Remember(key string, ttl time.Duration, target interface{}, f ResultFunc) error
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

func (c *RedisCache) Remember(key string, ttl time.Duration, target interface{}, f ResultFunc) error {
	str, err := c.store.Get(key).Result()
	if str != "" {
		if err := json.Unmarshal([]byte(str), &target); err != nil {
			c.log.Error(errors.Wrapf(err, "could parse payload from cache with key [%s]", key))
		} else {
			return nil
		}
	}

	v, err := f()
	if err != nil {
		return err
	}

	switch t := target.(type) {
	case *model.Microservice:
		*t =  *v.(*model.Microservice)
	case *model.Entity:
		*t = *v.(*model.Entity)
	default:
		panic("Cannot do nothing")
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



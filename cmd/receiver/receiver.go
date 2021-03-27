package main

import (
	"context"
	"github.com/denismitr/auditbase/internal/cache"
	"github.com/denismitr/auditbase/internal/receiver"
	"github.com/denismitr/auditbase/internal/utils"
	"github.com/denismitr/auditbase/internal/utils/clock"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denismitr/auditbase/internal/flow"
	"github.com/denismitr/auditbase/internal/queue"
	"github.com/denismitr/auditbase/internal/rest"
	"github.com/denismitr/auditbase/internal/utils/env"
	"github.com/denismitr/auditbase/internal/utils/logger"
	"github.com/pkg/profile"
)

func main() {
	env.LoadFromDotEnv()

	debug()

	lg := logger.NewStdoutLogger(env.StringOrDefault("APP_ENV", "prod"), "auditbase_rest_api")

	receiverAPI, err := create(lg)
	if err != nil {
		panic(err)
	}

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)

	lg.Debugf("All services are ready. Starting receiver...")
	stop := receiverAPI.Start()

	<-terminate

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := stop(ctx); err != nil {
		log.Error(err)
	}
}

func create(lg logger.Logger) (*rest.API, error) {
	startCtx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
	defer cancel()

	mq := queue.Rabbit(env.MustString("RABBITMQ_DSN"), lg, 3)

	if err := mq.Connect(startCtx); err != nil {
		return nil, err
	}

	af := flow.New(mq, lg, flow.NewConfigFromGlobals())

	if err := af.Scaffold(); err != nil {
		return nil, err
	}

	c := createRedisCache()

	restCfg := rest.Config{
		Port:      ":" + env.MustString("RECEIVER_API_PORT"),
		BodyLimit: "250K",
	}

	rc := receiver.New(lg, clock.New(), af, utils.NewUUID4Generator(), c)
	e := echo.New()
	return rest.NewReceiverAPI(e, restCfg, lg, rc), nil
}

func createRedisCache() *cache.RedisCache {
	c := redis.NewClient(&redis.Options{
		Addr:     env.MustString("REDIS_HOST") + ":" + env.MustString("REDIS_PORT"),
		Password: env.String("REDIS_PASSWORD"),
		DB:       env.IntOrDefault("REDIS_DB", 0),
	})

	if err := c.Ping().Err(); err != nil {
		panic(err)
	}

	return cache.NewRedisCache(c)
}

func debug() {
	if env.IsTruthy("APP_TRACE") {
		stopper := profile.Start(profile.CPUProfile, profile.ProfilePath("/tmp/debug/receiver"))

		go func() {
			ticker := time.After(2 * time.Minute)
			<-ticker
			stopper.Stop()
		}()
	}
}

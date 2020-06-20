package main

import (
	"context"
	"github.com/denismitr/auditbase/cache"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/rest"
	"github.com/denismitr/auditbase/utils/env"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/pkg/profile"
)

func main() {
	env.LoadFromDotEnv()

	debug()

	lg := logger.NewStdoutLogger(env.StringOrDefault("APP_ENV", "prod"), "auditbase_rest_api")

	receiver, err := create(lg)
	if err != nil {
		panic(err)
	}

	terminate := make(chan os.Signal)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)

	lg.Debugf("All services are ready. Starting receiver...")
	stop := receiver.Start()

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

	cacheCh := make(chan cache.Cacher)
	efCh := make(chan *flow.MQEventFlow)
	errCh := make(chan error)

	go func() {
		mq := queue.Rabbit(env.MustString("RABBITMQ_DSN"), lg, 3)

		if err := mq.Connect(startCtx); err != nil {
			errCh <- err
			return
		}

		ef := flow.New(mq, lg, flow.NewConfigFromGlobals())

		if err := ef.Scaffold(); err != nil {
			errCh <- err
			return
		}

		efCh <- ef
	}()

	go func() {
		opts := &redis.Options{
			Addr:     net.JoinHostPort(env.MustString("REDIS_HOST"), env.MustString("REDIS_PORT")),
			Password: env.String("REDIS_PASSWORD"),
			DB:       env.IntOrDefault("REDIS_DB", 0),
		}

		c, err := cache.ConnectRedis(startCtx, lg, opts)

		if err != nil {
			errCh <- err
			return
		}

		cacheCh <- c
	}()

	var ef *flow.MQEventFlow
	var cacher cache.Cacher
	var err error

	allServicesReady := func() bool {
		return ef != nil && cacher != nil
	}

done:
	for {
		select {
		case ef = <-efCh:
			if allServicesReady() {
				break done
			}
		case cacher = <-cacheCh:
			if allServicesReady() {
				break done
			}
		case err = <-errCh:
			break done
		}
	}

	close(efCh)
	close(errCh)
	close(cacheCh)

	if err != nil {
		return nil, err
	}

	restCfg := rest.Config{
		Port:      ":" + env.MustString("RECEIVER_API_PORT"),
		BodyLimit: "250K",
	}

	e := echo.New()

	return rest.NewReceiverAPI(e, restCfg, lg,  ef, cacher), nil
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

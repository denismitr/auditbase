package main

import (
	"context"
	"fmt"
	"github.com/denismitr/auditbase/cache"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denismitr/auditbase/db/mysql"
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/rest"
	"github.com/denismitr/auditbase/utils/env"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/profile"
)

func main() {
	env.LoadFromDotEnv()

	debug()

	fmt.Println("Waiting for DB connection...")
	time.Sleep(40 * time.Second)

	log := logger.NewStdoutLogger(env.StringOrDefault("APP_ENV", "prod"), "auditbase_rest_api")
	uuid4 := uuid.NewUUID4Generator()

	dbConn, err := sqlx.Connect("mysql", os.Getenv("AUDITBASE_DB_DSN"))
	if err != nil {
		panic(err)
	}

	migrator := mysql.Migrator(dbConn)

	if err := migrator.Up(); err != nil {
		panic(err)
	}

	factory := mysql.NewRepositoryFactory(dbConn, uuid4, log)

	mq := queue.Rabbit(env.MustString("RABBITMQ_DSN"), log, 4)

	startCtx, _ := context.WithTimeout(context.Background(), 60 * time.Second)
	if err := mq.Connect(startCtx); err != nil {
		panic(err)
	}

	flowCfg := flow.NewConfigFromGlobals()
	ef := flow.New(mq, log, flowCfg)

	if err := ef.Scaffold(); err != nil {
		panic(err)
	}

	restCfg := rest.Config{
		Port:      ":" + env.MustString("RECEIVER_API_PORT"),
		BodyLimit: "250K",
	}

	e := echo.New()
	cacher := connectRedis(log)

	receiver := rest.NewReceiverAPI(e, restCfg, log, factory, ef, cacher)

	terminate := make(chan os.Signal)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	stop := receiver.Start()

	<-terminate

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := stop(ctx); err != nil {
		log.Error(err)
	}
}

func connectRedis(log logger.Logger) *cache.RedisCache {
	c := redis.NewClient(&redis.Options{
		Addr:     env.MustString("REDIS_HOST") + ":" + env.MustString("REDIS_PORT"),
		Password: env.String("REDIS_PASSWORD"),
		DB:       env.IntOrDefault("REDIS_DB", 0),
	})

	if err := c.Ping().Err(); err != nil {
		panic(err)
	}

	return cache.NewRedisCache(c, log)
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

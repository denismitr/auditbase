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
	time.Sleep(20 * time.Second)

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

	microservices := mysql.NewMicroserviceRepository(dbConn, uuid4)
	events := mysql.NewEventRepository(dbConn, uuid4)
	entities := mysql.NewEntityRepository(dbConn, uuid4, log)

	mq := queue.NewRabbitQueue(env.MustString("RABBITMQ_DSN"), log, 4)

	if err := mq.Connect(); err != nil {
		panic(err)
	}

	port := ":" + env.MustString("BACK_OFFICE_API_PORT")

	flowCfg := flow.NewConfigFromGlobals()
	ef := flow.New(mq, log, flowCfg)

	if err := ef.Scaffold(); err != nil {
		panic(err)
	}

	restCfg := rest.Config{
		Port:      port,
		BodyLimit: "250K",
	}

	e := echo.New()
	cacher := connectRedis(log)

	backOffice := rest.NewBackOfficeAPI(
		e,
		restCfg,
		log,
		ef,
		microservices,
		events,
		entities,
		cacher,
	)

	terminate := make(chan os.Signal)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	stop := backOffice.Start()

	<-terminate

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := stop(ctx); err != nil {
		log.Error(err)
	}
}

func debug() {
	if env.IsTruthy("APP_TRACE") {
		stopper := profile.Start(profile.CPUProfile, profile.ProfilePath("/tmp/debug/backoffice"))

		go func() {
			ticker := time.After(2 * time.Minute)
			<-ticker
			stopper.Stop()
		}()
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

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/denismitr/auditbase/cache"
	"github.com/go-redis/redis/v7"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denismitr/auditbase/consumer"
	"github.com/denismitr/auditbase/db"
	"github.com/denismitr/auditbase/db/mysql"
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/utils/env"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/pkg/profile"
)

const defaultConsumerName = "auditbase_consumer"
const defaultErrorsConsumerName = "auditbase_requeue_consumer"

func main() {
	var errorsConsumer = flag.Bool("errors", false, "Consumer that consumes requeued messages")
	var name = flag.String("name", defaultConsumerName, "Consumer name")

	flag.Parse()

	env.LoadFromDotEnv()
	cfg := flow.NewConfigFromGlobals()

	queueName, consumerName := resolveNames(*errorsConsumer, cfg, *name)
	log := logger.NewStdoutLogger(env.StringOrDefault("APP_ENV", "prod"), consumerName)

	debug(*errorsConsumer)

	run(log, cfg, consumerName, queueName)
}

func resolveNames(errorsConsumer bool, cfg flow.Config, consumerName string) (string, string) {
	var queueName string

	if errorsConsumer == true {
		queueName = cfg.ErrorQueueName
		if consumerName == defaultConsumerName {
			consumerName = defaultErrorsConsumerName
		}
	} else {
		queueName = cfg.QueueName
	}

	return queueName, consumerName
}

func run(log logger.Logger, cfg flow.Config, consumerName, queueName string) {
	fmt.Println("Waiting for DB connection")
	time.Sleep(20 * time.Second)

	uuid4 := uuid.NewUUID4Generator()

	dbConn, err := sqlx.Connect("mysql", env.MustString("AUDITBASE_DB_DSN"))
	if err != nil {
		panic(err)
	}

	dbConn.SetMaxOpenConns(100)

	if err := mysql.Migrator(dbConn).Up(); err != nil {
		panic(err)
	}

	microservices := mysql.NewMicroserviceRepository(dbConn, uuid4)
	events := mysql.NewEventRepository(dbConn, uuid4)
	entities := mysql.NewEntityRepository(dbConn, uuid4, log)

	cacher := connectRedis(log)
	persister := db.NewDBPersister(microservices, events, entities, log, cacher)
	mq := queue.NewRabbitQueue(env.MustString("RABBITMQ_DSN"), log, 3)

	if err := mq.Connect(); err != nil {
		panic(err)
	}

	ef := flow.New(mq, log, cfg)

	if err := ef.Scaffold(); err != nil {
		panic(err)
	}

	c := consumer.New(ef, log, persister)

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	stop := c.Start(queueName, consumerName)

	<-terminate
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
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

func debug(isErrorsConsumer bool) {
	if env.IsTruthy("APP_TRACE") && isErrorsConsumer == false {
		stopper := profile.Start(profile.CPUProfile, profile.MemProfile, profile.ProfilePath("/tmp/debug/consumer"))

		go func() {
			ticker := time.After(2 * time.Minute)
			<-ticker
			stopper.Stop()
		}()
	}
}

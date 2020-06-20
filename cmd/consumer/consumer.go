package main

import (
	"context"
	"flag"
	"github.com/denismitr/auditbase/cache"
	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/gommon/log"
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
	lg := logger.NewStdoutLogger(env.StringOrDefault("APP_ENV", "prod"), consumerName)

	debug(*errorsConsumer)

	run(lg, cfg, consumerName, queueName)
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

func create(lg logger.Logger, cfg flow.Config) (*consumer.Consumer, error) {
	startCtx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
	defer cancel()

	connCh := make(chan *sqlx.DB)
	cacheCh := make(chan cache.Cacher)
	efCh := make(chan *flow.MQEventFlow)
	errCh := make(chan error)

	go func() {
		dbConn, err := mysql.ConnectAndMigrate(startCtx, lg, env.MustString("AUDITBASE_DB_DSN"), 150)
		if err != nil {
			errCh <- err
			return
		}

		connCh <- dbConn
	}()

	go func() {
		mq := queue.Rabbit(env.MustString("RABBITMQ_DSN"), lg, 3)

		if err := mq.Connect(startCtx); err != nil {
			errCh <- err
			return
		}

		ef := flow.New(mq, lg, cfg)

		if err := ef.Scaffold(); err != nil {
			errCh <- err
			return
		}

		efCh <- ef
	}()

	go func() {
		opts := &redis.Options{
			Addr:     env.MustString("REDIS_HOST") + ":" + env.MustString("REDIS_PORT"),
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

	var conn *sqlx.DB
	var ef *flow.MQEventFlow
	var cacher cache.Cacher
	var err error

	allServicesReady := func() bool {
		return conn != nil && ef != nil && cacher != nil
	}

done:
	for {
		select {
			case ef = <-efCh:
				if allServicesReady() {
					break done
				}
			case conn = <-connCh:
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
	close(connCh)
	close(errCh)
	close(cacheCh)

	if err != nil {
		return nil, err
	}

	uuid4 := uuid.NewUUID4Generator()
	factory := mysql.NewRepositoryFactory(conn, uuid4, lg)
	persister := db.NewDBPersister(factory, lg, cacher)

	return consumer.New(ef, lg, persister), nil
}

func run(lg logger.Logger, cfg flow.Config, consumerName, queueName string) {
	c, err := create(lg, cfg)
	if err != nil {
		panic(err)
	}

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

package main

import (
	"context"
	"flag"
	"fmt"
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
const defaultRequeueConsumerName = "auditbase_requeue_consumer"

func main() {
	var requeueConsumer = flag.Bool("requeued", false, "Consumer that consumes requeued messages")
	var consumerName = flag.String("name", defaultConsumerName, "Consumer name")
	var queueName string

	flag.Parse()

	env.LoadFromDotEnv()
	cfg := flow.NewConfigFromGlobals()

	debug(*requeueConsumer)

	if *requeueConsumer == true {
		queueName = cfg.ErrorQueueName
		if *consumerName == defaultConsumerName {
			*consumerName = defaultRequeueConsumerName
		}
	} else {
		queueName = cfg.QueueName
	}

	logger := logger.NewStdoutLogger(env.StringOrDefault("APP_ENV", "prod"), *consumerName)

	run(logger, cfg, *consumerName, queueName)
}

func run(logger logger.Logger, cfg flow.Config, consumerName, queueName string) {
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
	entities := mysql.NewEntityRepository(dbConn, uuid4)
	persister := db.NewDBPersister(microservices, events, entities, logger)
	mq := queue.NewRabbitQueue(env.MustString("RABBITMQ_DSN"), logger, 3)

	if err := mq.Connect(); err != nil {
		panic(err)
	}

	ef := flow.New(mq, logger, cfg)

	if err := ef.Scaffold(); err != nil {
		panic(err)
	}

	consumer := consumer.New(ef, logger, persister)

	ctx, cancel := context.WithCancel(context.Background())

	terminate := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)

	consumer.Start(ctx, queueName, consumerName)

	go func() {
		<-terminate
		cancel()
		close(done)
	}()

	<-done
}

func debug(isRequeueConsumer bool) {
	if env.IsTruthy("APP_TRACE") && isRequeueConsumer == false {
		stopper := profile.Start(profile.CPUProfile, profile.MemProfile, profile.ProfilePath("/tmp/debug/consumer"))

		go func() {
			ticker := time.After(2 * time.Minute)
			<-ticker
			stopper.Stop()
		}()
	}
}

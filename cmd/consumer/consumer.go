package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denismitr/auditbase/consumer"
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/sql/mysql"
	"github.com/denismitr/auditbase/utils"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

const defaultConsumerName = "auditbase_consumer"
const defaultRequeueConsumerName = "auditbase_requeue_consumer"

func main() {
	var requeueConsumer = flag.Bool("requeued", false, "Consumer that consumes requeued messages")
	var consumerName = flag.String("name", defaultConsumerName, "Consumer name")
	var queueName string

	flag.Parse()
	loadEnvVars()
	cfg := flow.NewConfigFromGlobals()

	if *requeueConsumer == true {
		queueName = cfg.ErrorQueueName
		if *consumerName == defaultConsumerName {
			*consumerName = defaultRequeueConsumerName
		}
	} else {
		queueName = cfg.QueueName
	}

	logger := utils.NewStdoutLogger(os.Getenv("APP_ENV"), *consumerName)

	run(logger, cfg, *consumerName, queueName)
}

func run(logger utils.Logger, cfg flow.Config, consumerName, queueName string) {
	fmt.Println("Waiting for DB connection")
	time.Sleep(20 * time.Second)

	uuid4 := utils.NewUUID4Generator()

	dbConn, err := sqlx.Connect("mysql", os.Getenv("AUDITBASE_DB_DSN"))
	if err != nil {
		panic(err)
	}

	dbConn.SetMaxOpenConns(100)

	microservices := mysql.NewMicroserviceRepository(dbConn, uuid4)
	events := mysql.NewEventRepository(dbConn, uuid4)
	targetTypes := mysql.NewTargetTypeRepository(dbConn, uuid4)
	actorTypes := mysql.NewActorTypeRepository(dbConn, uuid4)

	mq := queue.NewRabbitQueue(os.Getenv("RABBITMQ_DSN"), logger, 3)

	if err := mq.Connect(); err != nil {
		panic(err)
	}

	ef := flow.New(mq, logger, cfg)

	if err := ef.Scaffold(); err != nil {
		panic(err)
	}

	consumer := consumer.New(
		ef,
		logger,
		mq,
		microservices,
		events,
		targetTypes,
		actorTypes,
	)

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

func loadEnvVars() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

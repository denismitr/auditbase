package main

import (
	"context"
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

func main() {
	loadEnvVars()

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

	logger := utils.NewStdoutLogger(os.Getenv("APP_ENV"), "auditbase_consumer")
	mq := queue.NewRabbitQueue(os.Getenv("RABBITMQ_DSN"), logger, 3)

	if err := mq.Connect(); err != nil {
		panic(err)
	}

	cfg := flow.NewConfigFromGlobals()
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

	consumer.Start(ctx, "event_consumer")

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

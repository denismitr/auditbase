package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
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
	time.Sleep(20)

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
	mq.WaitForConnection()

	exchange := os.Getenv("EVENTS_EXCHANGE")
	routingKey := os.Getenv("EVENTS_ROUTING_KEY")
	queueName := os.Getenv("EVENTS_QUEUE_NAME")
	exchangeType := os.Getenv("EVENTS_EXCHANGE_TYPE")

	cfg := flow.NewConfig(exchange, exchangeType, routingKey, queueName, true)
	ef := flow.NewMQEventFlow(mq, cfg)

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

	quit := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(quit, os.Interrupt)

	stop := consumer.Start("event_consumer")

	// TODO: fix context cancel
	go gracefulShutdown(quit, done, stop)

	<-done
}

func gracefulShutdown(quit chan os.Signal, done chan struct{}, stop consumer.StopFunc) {
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	stop(ctx)

	close(done)

	fmt.Println("Graceful shutdown is done")
}

func loadEnvVars() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/denismitr/auditbase/consumer"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/sql/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	loadEnvVars()

	fmt.Println("Waiting for DB connection")
	time.Sleep(20)

	dbConn, err := sqlx.Connect("mysql", os.Getenv("AUDITBASE_DB_DSN"))
	if err != nil {
		panic(err)
	}

	microservices := &mysql.MicroserviceRepository{Conn: dbConn}
	events := &mysql.EventRepository{Conn: dbConn}
	targetTypes := &mysql.TargetTypeRepository{Conn: dbConn}
	actorTypes := &mysql.ActorTypeRepository{Conn: dbConn}

	logger := logrus.New()
	mq := queue.NewRabbitQueue(os.Getenv("RABBITMQ_DSN"), logger, 3)
	mq.WaitForConnection()

	exchange := os.Getenv("EVENTS_EXCHANGE")
	routingKey := os.Getenv("EVENTS_ROUTING_KEY")
	queueName := os.Getenv("EVENTS_QUEUE_NAME")

	ee := &queue.DirectEventExchange{
		MQ:         mq,
		Exchange:   exchange,
		RoutingKey: routingKey,
		QueueName:  queueName,
	}

	consumer := consumer.New(logger, ee, microservices, events, targetTypes, actorTypes)

	consumer.Start()
}

func loadEnvVars() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

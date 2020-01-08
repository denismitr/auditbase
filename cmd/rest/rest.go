package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/rest"
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

	if err := mysql.Scaffold(dbConn); err != nil {
		panic(err)
	}

	microservices := &mysql.MicroserviceRepository{Conn: dbConn}
	events := &mysql.EventRepository{Conn: dbConn}

	logger := logrus.New()
	mq := queue.NewRabbitQueue(os.Getenv("RABBITMQ_DSN"), logger, 3)
	mq.WaitForConnection()

	exchange := os.Getenv("EVENTS_EXCHANGE")
	routingKey := os.Getenv("EVENTS_ROUTING_KEY")
	queueName := os.Getenv("EVENTS_QUEUE_NAME")
	port := ":" + os.Getenv("REST_API_PORT")

	d := queue.NewDelivery(queueName, exchange, routingKey, "direct", true)
	ee := queue.NewDirectEventExchange(mq, d)

	if err := ee.Scaffold(); err != nil {
		panic(err)
	}

	rest := rest.New(rest.Config{
		Port: port,
	}, ee, microservices, events)

	rest.Start()
}

func loadEnvVars() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

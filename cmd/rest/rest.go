package main

import (
	"fmt"
	"time"

	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/rest"
	"github.com/denismitr/auditbase/sql/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("Waiting for DB connection")
	time.Sleep(10)

	dbConn, err := sqlx.Connect("mysql", "auditbase:secret@(auditbase_db:3306)/auditbase")
	if err != nil {
		panic(err)
	}

	mysqlRepo := &mysql.MicroserviceRepository{Conn: dbConn}

	logger := logrus.New()
	queue := queue.NewRabbitQueue("amqp://auditbase:secret@auditbase_rabbit:5672/", logger, 3)
	queue.WaitForConnection()
	rest := rest.New(rest.Config{
		Port: ":3000",
	}, queue, mysqlRepo)

	rest.Start()
}

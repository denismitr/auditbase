package main

import (
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/rest"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	queue := queue.NewRabbitQueue("amqp://auditbase:secret@auditbase_rabbit:5672/", logger, 3)
	queue.WaitForConnection()
	rest := rest.New(rest.Config{
		Port: ":3000",
	}, queue)

	rest.Start()
}

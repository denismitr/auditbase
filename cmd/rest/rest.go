package main

import (
	"fmt"
	"os"
	"time"

	"github.com/denismitr/auditbase/db/mysql"
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/rest"
	"github.com/denismitr/auditbase/utils/env"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/profile"
)

func main() {
	env.LoadFromDotEnv()

	debug()

	fmt.Println("Waiting for DB connection...")
	time.Sleep(20 * time.Second)

	uuid4 := uuid.NewUUID4Generator()

	dbConn, err := sqlx.Connect("mysql", os.Getenv("AUDITBASE_DB_DSN"))
	if err != nil {
		panic(err)
	}

	migrator := mysql.Migrator(dbConn)

	if err := migrator.Up(); err != nil {
		panic(err)
	}

	microservices := mysql.NewMicroserviceRepository(dbConn, uuid4)
	events := mysql.NewEventRepository(dbConn, uuid4)
	entities := mysql.NewEntityRepository(dbConn, uuid4)

	logger := logger.NewStdoutLogger(env.StringOrDefault("APP_ENV", "prod"), "auditbase_rest_api")
	mq := queue.NewRabbitQueue(env.MustString("RABBITMQ_DSN"), logger, 4)

	if err := mq.Connect(); err != nil {
		panic(err)
	}

	port := ":" + env.MustString("REST_API_PORT")

	flowCfg := flow.NewConfigFromGlobals()
	ef := flow.New(mq, logger, flowCfg)

	if err := ef.Scaffold(); err != nil {
		panic(err)
	}

	restCfg := rest.Config{
		Port:      port,
		BodyLimit: "250K",
	}

	rest := rest.New(
		restCfg,
		logger,
		ef,
		microservices,
		events,
		entities,
	)

	rest.Start()
}

func debug() {
	if env.IsTruthy("APP_TRACE") {
		stopper := profile.Start(profile.CPUProfile, profile.ProfilePath("/tmp/debug/rest"))

		go func() {
			ticker := time.After(2 * time.Minute)
			<-ticker
			stopper.Stop()
		}()
	}
}

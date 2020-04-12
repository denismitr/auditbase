package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/denismitr/auditbase/db/mysql"
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/rest"
	"github.com/denismitr/auditbase/utils"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/pkg/profile"
)

func main() {
	loadEnvVars()

	debug()

	fmt.Println("Waiting for DB connection...")
	time.Sleep(20 * time.Second)

	uuid4 := utils.NewUUID4Generator()

	dbConn, err := sqlx.Connect("mysql", os.Getenv("AUDITBASE_DB_DSN"))
	if err != nil {
		panic(err)
	}

	if err := mysql.Scaffold(dbConn); err != nil {
		panic(err)
	}

	microservices := mysql.NewMicroserviceRepository(dbConn, uuid4)
	events := mysql.NewEventRepository(dbConn, uuid4)
	actorTypes := mysql.NewActorTypeRepository(dbConn, uuid4)
	targetTypes := mysql.NewTargetTypeRepository(dbConn, uuid4)

	logger := utils.NewStdoutLogger(os.Getenv("APP_ENV"), "auditbase_rest_api")
	mq := queue.NewRabbitQueue(os.Getenv("RABBITMQ_DSN"), logger, 4)

	if err := mq.Connect(); err != nil {
		panic(err)
	}

	port := ":" + os.Getenv("REST_API_PORT")

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
		actorTypes,
		targetTypes,
	)

	rest.Start()
}

func loadEnvVars() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

func debug() {
	if os.Getenv("APP_TRACE") != "" && os.Getenv("APP_TRACE") != "0" {
		stopper := profile.Start(profile.CPUProfile, profile.ProfilePath("/tmp/debug/rest"))

		go func() {
			ticker := time.After(2 * time.Minute)
			<-ticker
			stopper.Stop()
		}()
	}
}

package main

import (
	"context"
	"github.com/denismitr/auditbase/service"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/denismitr/auditbase/db/mysql"
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/rest"
	"github.com/denismitr/auditbase/utils/env"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/profile"
)

func main() {
	env.LoadFromDotEnv()

	debug(env.IsTruthy("APP_TRACE"))

	lg := logger.NewStdoutLogger(env.StringOrDefault("APP_ENV", "prod"), "auditbase_rest_api")

	port := ":" + env.MustString("BACK_OFFICE_API_PORT")

	restCfg := rest.Config{
		Port:      port,
		BodyLimit: "250K",
	}

	backOffice, err := createBackOffice(lg, restCfg)
	if err != nil {
		panic(err)
	}

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)

	stop := backOffice.Start()

	<-terminate

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := stop(ctx); err != nil {
		log.Error(err)
	}
}

func debug(run bool) {
	if run {
		stopper := profile.Start(profile.CPUProfile, profile.ProfilePath("/tmp/debug/backoffice"))

		go func() {
			ticker := time.After(2 * time.Minute)
			<-ticker
			stopper.Stop()
		}()
	}
}


func createBackOffice(lg logger.Logger, restCfg rest.Config) (*rest.API, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	connCh := make(chan *sqlx.DB, 1)
	afCh := make(chan *flow.MQActionFlow, 1)
	errCh := make(chan error, 2)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		mq := queue.Rabbit(env.MustString("RABBITMQ_DSN"), lg, 3)

		if err := mq.Connect(ctx); err != nil {
			errCh <- err
			return
		}

		ef := flow.New(mq, lg, flow.NewConfigFromGlobals())

		if err := ef.Scaffold(); err != nil {
			errCh <- err
			return
		}

		afCh <- ef
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		dbConn, err := mysql.ConnectAndMigrate(ctx, lg, env.MustString("AUDITBASE_DB_DSN"), 50, 10)
		if err != nil {
			errCh <- err
			return
		}

		connCh <- dbConn
	}()

	wg.Wait()

	select {
	case err :=  <-errCh:
		return nil, err
	default:
		lg.Debugf("Connection to DB and RabbitMQ have been established")
	}

	db := mysql.NewDatabase(<-connCh, lg)

	services := rest.BackOfficeServices{
		Actions: service.NewActionService(db, lg),
		Microservices: service.NewMicroserviceService(db, lg),
		Entities: service.NewEntityService(db, lg),
	}

	return rest.BackOfficeAPI(echo.New(), restCfg, lg, <-afCh, services), nil
}

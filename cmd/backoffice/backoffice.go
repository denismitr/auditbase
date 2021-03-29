package main

import (
	"context"
	"github.com/denismitr/auditbase/internal/service"
	"github.com/denismitr/auditbase/internal/utils/env"
	"github.com/denismitr/goenv"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/denismitr/auditbase/internal/db/mysql"
	"github.com/denismitr/auditbase/internal/flow"
	"github.com/denismitr/auditbase/internal/flow/queue"
	"github.com/denismitr/auditbase/internal/rest"
	"github.com/denismitr/auditbase/internal/utils/logger"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/profile"
)

func main() {
	env.LoadFromDotEnv()

	debug(goenv.IsTruthy("APP_TRACE"))

	lg := logger.NewStdoutLogger(goenv.StringOrDefault("APP_ENV", "prod"), "auditbase_rest_api")

	port := ":" + goenv.MustString("BACK_OFFICE_API_PORT")

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

		mq := queue.Rabbit(goenv.MustString("RABBITMQ_DSN"), lg, 3)

		if err := mq.Connect(ctx); err != nil {
			errCh <- err
			return
		}

		ef := flow.New(mq, lg, flow.Config{
			ExchangeName: goenv.MustString("ACTIONS_EXCHANGE"),
			ActionsCreateQueue: goenv.MustString("NEW_ACTIONS_QUEUE"),
			ActionsUpdateQueue: goenv.MustString("UPDATE_ACTIONS_QUEUE"),
			Concurrency: goenv.IntOrDefault("CONSUMER_CONCURRENCY", 4),
			ExchangeType: goenv.MustString("ACTIONS_EXCHANGE_TYPE"),
			MaxRequeue: goenv.IntOrDefault("ACTIONS_MAX_REQUEUE", 2),
			IsPeristent: true,
		})

		if err := ef.Scaffold(); err != nil {
			errCh <- err
			return
		}

		afCh <- ef
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		dbConn, err := mysql.ConnectAndMigrate(ctx, lg, goenv.MustString("AUDITBASE_DB_DSN"), 50, 10)
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
		Actions:       service.NewActionService(db, lg),
		Microservices: service.NewMicroserviceService(db, lg),
		Entities:      service.NewEntityService(db, lg),
	}

	return rest.BackOfficeAPI(echo.New(), restCfg, lg, <-afCh, services), nil
}

package main

import (
	"context"
	"github.com/denismitr/auditbase/internal/service"
	"github.com/denismitr/goenv"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/denismitr/auditbase/internal/consumer"
	"github.com/denismitr/auditbase/internal/db/mysql"
	"github.com/denismitr/auditbase/internal/flow"
	"github.com/denismitr/auditbase/internal/flow/queue"
	"github.com/denismitr/auditbase/internal/utils/env"
	"github.com/denismitr/auditbase/internal/utils/logger"
	"github.com/pkg/profile"
)

const defaultConsumerName = "actions_consumer"

func main() {
	env.LoadFromDotEnv()

	lg := logger.NewStdoutLogger(goenv.MustString("APP_ENV"), "ACTIONS_CONSUMER")

	if err := run(lg); err != nil {
		panic(err)
	}
}

func run(lg logger.Logger) error {
	cfg := flow.Config{
		ExchangeName: goenv.MustString("ACTIONS_EXCHANGE"),
		ActionsCreateQueue: goenv.MustString("NEW_ACTIONS_QUEUE"),
		ActionsUpdateQueue: goenv.MustString("UPDATE_ACTIONS_QUEUE"),
		Concurrency: goenv.IntOrDefault("CONSUMER_CONCURRENCY", 4),
		ExchangeType: goenv.MustString("ACTIONS_EXCHANGE_TYPE"),
		MaxRequeue: goenv.IntOrDefault("ACTIONS_MAX_REQUEUE", 2),
		IsPeristent: true,
	}

	consumerName := goenv.StringOrDefault("CONSUMER_NAME", defaultConsumerName)

	c, err := createConsumer(consumerName, lg, cfg)
	if err != nil {
		return err
	}

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)
	errCh := make(chan error, 1)
	doneCh := make(chan struct{})

	go func() {
		err := <- c.Start(terminate)
		if err != nil {
			lg.Error(err)
			os.Exit(1)
		} else {
			doneCh <- struct{}{}
		}
	}()

	for {
		select {
			case err := <-errCh:
				lg.Error(errors.Errorf("consumer [%s] exiting with error %s", err))
				return err
			case <-doneCh:
				lg.Error(errors.Errorf("consumer [%s] is done", consumerName))
				return nil
		}
	}
}

func createConsumer(consumerName string, lg logger.Logger, cfg flow.Config) (*consumer.Consumer, error) {
	startCtx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
	defer cancel()

	connCh := make(chan *sqlx.DB, 1)
	afCh := make(chan *flow.MQActionFlow, 1)
	errCh := make(chan error, 2)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		dbConn, err := mysql.ConnectAndMigrate(startCtx, lg, goenv.MustString("AUDITBASE_DB_DSN"), 200, 20)
		if err != nil {
			errCh <- err
			return
		}

		connCh <- dbConn
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		mq := queue.Rabbit(goenv.MustString("RABBITMQ_DSN"), lg, 3)

		if err := mq.Connect(startCtx); err != nil {
			errCh <- err
			return
		}

		af := flow.New(mq, lg, cfg)

		if err := af.Scaffold(); err != nil {
			errCh <- err
			return
		}

		afCh <- af
	}()

	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
		lg.Debugf("consumer dependencies activated")
	}

	conn := <-connCh
	af := <-afCh

	db := mysql.NewDatabase(conn, lg)
	actionService := service.NewActionService(db, lg)

	return consumer.New(consumerName, af, lg, actionService), nil
}

func debug(isErrorsConsumer bool) {
	if goenv.IsTruthy("APP_TRACE") && !isErrorsConsumer {
		stopper := profile.Start(profile.CPUProfile, profile.MemProfile, profile.ProfilePath("/tmp/debug/consumer"))

		go func() {
			ticker := time.After(2 * time.Minute)
			<-ticker
			stopper.Stop()
		}()
	}
}

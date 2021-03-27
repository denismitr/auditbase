package main

import (
	"context"
	"flag"
	"github.com/denismitr/auditbase/internal/service"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denismitr/auditbase/internal/consumer"
	"github.com/denismitr/auditbase/internal/db/mysql"
	"github.com/denismitr/auditbase/internal/flow"
	"github.com/denismitr/auditbase/internal/queue"
	"github.com/denismitr/auditbase/internal/utils/env"
	"github.com/denismitr/auditbase/internal/utils/logger"
	"github.com/pkg/profile"
)

const defaultConsumerName = "auditbase_consumer"
const defaultErrorsConsumerName = "auditbase_requeue_consumer"

func main() {
	var errorsConsumer = flag.Bool("errors", false, "Consumer that consumes requeued messages")
	var name = flag.String("name", defaultConsumerName, "Consumer name")

	flag.Parse()

	env.LoadFromDotEnv()
	cfg := flow.NewConfigFromGlobals()

	queueName, consumerName := resolveNames(*errorsConsumer, cfg, *name)
	lg := logger.NewStdoutLogger(env.StringOrDefault("APP_ENV", "prod"), consumerName)

	debug(*errorsConsumer)

	if err := run(lg, cfg, consumerName, queueName); err != nil {
		panic(err)
	}
}

func resolveNames(errorsConsumer bool, cfg flow.Config, consumerName string) (string, string) {
	var queueName string

	if errorsConsumer {
		queueName = cfg.ErrorQueueName
		if consumerName == defaultConsumerName {
			consumerName = defaultErrorsConsumerName
		}
	} else {
		queueName = cfg.QueueName
	}

	return queueName, consumerName
}

func createConsumer(lg logger.Logger, cfg flow.Config) (*consumer.Consumer, error) {
	startCtx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
	defer cancel()

	connCh := make(chan *sqlx.DB)
	efCh := make(chan *flow.MQActionFlow)
	errCh := make(chan error)

	go func() {
		dbConn, err := mysql.ConnectAndMigrate(startCtx, lg, env.MustString("AUDITBASE_DB_DSN"), 200, 20)
		if err != nil {
			errCh <- err
			return
		}

		connCh <- dbConn
	}()

	go func() {
		mq := queue.Rabbit(env.MustString("RABBITMQ_DSN"), lg, 3)

		if err := mq.Connect(startCtx); err != nil {
			errCh <- err
			return
		}

		ef := flow.New(mq, lg, cfg)

		if err := ef.Scaffold(); err != nil {
			errCh <- err
			return
		}

		efCh <- ef
	}()

	var conn *sqlx.DB
	var ef *flow.MQActionFlow
	var err error

	allServicesReady := func() bool {
		return conn != nil && ef != nil
	}

done:
	for {
		select {
			case ef = <-efCh:
				if allServicesReady() {
					break done
				}
			case conn = <-connCh:
				if allServicesReady() {
					break done
				}
			case err = <-errCh:
				break done
		}
	}

	close(efCh)
	close(connCh)
	close(errCh)

	if err != nil {
		return nil, err
	}

	db := mysql.NewDatabase(conn, lg)
	actionService := service.NewActionService(db, lg)

	return consumer.New(ef, lg, actionService), nil
}

func run(lg logger.Logger, cfg flow.Config, consumerName, queueName string) error {
	c, err := createConsumer(lg, cfg)
	if err != nil {
		panic(err)
	}

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)
	errCh := make(chan error, 1)
	doneCh := make(chan struct{})

	go func() {
		if err := c.Start(terminate, queueName, consumerName); err != nil {
			errCh <- err
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

func debug(isErrorsConsumer bool) {
	if env.IsTruthy("APP_TRACE") && !isErrorsConsumer {
		stopper := profile.Start(profile.CPUProfile, profile.MemProfile, profile.ProfilePath("/tmp/debug/consumer"))

		go func() {
			ticker := time.After(2 * time.Minute)
			<-ticker
			stopper.Stop()
		}()
	}
}

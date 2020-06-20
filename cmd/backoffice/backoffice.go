package main

import (
	"context"
	"github.com/denismitr/auditbase/cache"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"net"
	"os"
	"os/signal"
	"syscall"
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

	debug(env.IsTruthy("APP_TRACE"))

	lg := logger.NewStdoutLogger(env.StringOrDefault("APP_ENV", "prod"), "auditbase_rest_api")

	port := ":" + env.MustString("BACK_OFFICE_API_PORT")

	restCfg := rest.Config{
		Port:      port,
		BodyLimit: "250K",
	}

	backOffice, err := create(lg, restCfg)
	if err != nil {
		panic(err)
	}

	terminate := make(chan os.Signal)
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


func create(lg logger.Logger, restCfg rest.Config) (*rest.API, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	connCh := make(chan *sqlx.DB)
	cacheCh := make(chan cache.Cacher)
	efCh := make(chan *flow.MQEventFlow)
	errCh := make(chan error)

	go func() {
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

		efCh <- ef
	}()

	go func() {
		opts := &redis.Options{
			Addr:     net.JoinHostPort(env.MustString("REDIS_HOST"), env.MustString("REDIS_PORT")),
			Password: env.String("REDIS_PASSWORD"),
			DB:       env.IntOrDefault("REDIS_DB", 0),
		}

		c, err := cache.ConnectRedis(ctx, lg, opts)

		if err != nil {
			errCh <- err
			return
		}

		cacheCh <- c
	}()

	go func() {
		dbConn, err := mysql.ConnectAndMigrate(ctx, lg, env.MustString("AUDITBASE_DB_DSN"), 150)
		if err != nil {
			errCh <- err
			return
		}

		connCh <- dbConn
	}()

	var conn *sqlx.DB
	var ef *flow.MQEventFlow
	var cacher cache.Cacher
	var err error

	allServicesReady := func() bool {
		return conn != nil && ef != nil && cacher != nil
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
		case cacher = <-cacheCh:
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
	close(cacheCh)

	if err != nil {
		return nil, err
	}

	return rest.BackOfficeAPI(
		echo.New(),
		restCfg,
		lg,
		ef,
		mysql.Factory(conn, uuid.NewUUID4Generator(), lg),
		cacher,
	), nil
}

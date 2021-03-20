package main

import (
	"context"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/queue"
	"github.com/denismitr/auditbase/rest"
	"github.com/denismitr/auditbase/utils/env"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/pkg/profile"
)

func main() {
	env.LoadFromDotEnv()

	debug()

	lg := logger.NewStdoutLogger(env.StringOrDefault("APP_ENV", "prod"), "auditbase_rest_api")

	receiver, err := create(lg)
	if err != nil {
		panic(err)
	}

	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)

	lg.Debugf("All services are ready. Starting receiver...")
	stop := receiver.Start()

	<-terminate

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := stop(ctx); err != nil {
		log.Error(err)
	}
}

func create(lg logger.Logger) (*rest.API, error) {
	startCtx, cancel := context.WithTimeout(context.Background(), 60 * time.Second)
	defer cancel()

	mq := queue.Rabbit(env.MustString("RABBITMQ_DSN"), lg, 3)

	if err := mq.Connect(startCtx); err != nil {
		return nil, err
	}

	af := flow.New(mq, lg, flow.NewConfigFromGlobals())

	if err := af.Scaffold(); err != nil {
		return nil, err
	}

	restCfg := rest.Config{
		Port:      ":" + env.MustString("RECEIVER_API_PORT"),
		BodyLimit: "250K",
	}

	e := echo.New()

	return rest.NewReceiverAPI(e, restCfg, lg,  af), nil
}

func debug() {
	if env.IsTruthy("APP_TRACE") {
		stopper := profile.Start(profile.CPUProfile, profile.ProfilePath("/tmp/debug/receiver"))

		go func() {
			ticker := time.After(2 * time.Minute)
			<-ticker
			stopper.Stop()
		}()
	}
}

package main

import (
	"context"
	"github.com/denismitr/auditbase/cache"
	"github.com/denismitr/auditbase/utils/retry"
	"github.com/go-redis/redis/v7"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/labstack/gommon/log"
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

	startCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	backOffice, err := createBackOffice(startCtx)
	cancel()
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

func connectRedis(log logger.Logger) *cache.RedisCache {
	c := redis.NewClient(&redis.Options{
		Addr:     env.MustString("REDIS_HOST") + ":" + env.MustString("REDIS_PORT"),
		Password: env.String("REDIS_PASSWORD"),
		DB:       env.IntOrDefault("REDIS_DB", 0),
	})

	if err := c.Ping().Err(); err != nil {
		panic(err)
	}

	return cache.NewRedisCache(c, log)
}

func createBackOffice(ctx context.Context) (*rest.API, error) {
	lg := logger.NewStdoutLogger(env.StringOrDefault("APP_ENV", "prod"), "auditbase_rest_api")

	var dbConn *sqlx.DB
	if err := retry.Incremental(ctx, 2 * time.Second, 100, func(attempt int) (err error) {
		dbConn, err = sqlx.Connect("mysql", os.Getenv("AUDITBASE_DB_DSN"))
		if err != nil {
			lg.Debugf("could not connect to DB on attempt %d", attempt)
			return retry.Error(err, attempt)
		}

		if _, err = dbConn.QueryxContext(ctx, "select 1"); err != nil {
			lg.Debugf("could not ping DB connection on attempt %d", attempt)
			return retry.Error(err, attempt)
		}

		lg.Debugf("connection with DB established")
		return nil
	}); err != nil {
		return nil, err
	}


	uuid4 := uuid.NewUUID4Generator()

	migrator := mysql.Migrator(dbConn)

	if err := migrator.Up(); err != nil {
		panic(err)
	}

	factory := mysql.NewRepositoryFactory(dbConn, uuid4, lg)

	mq := queue.Rabbit(env.MustString("RABBITMQ_DSN"), lg, 30)

	if err := mq.Connect(ctx); err != nil {
		return nil, err
	}

	port := ":" + env.MustString("BACK_OFFICE_API_PORT")

	flowCfg := flow.NewConfigFromGlobals()
	ef := flow.New(mq, lg, flowCfg)

	if err := ef.Scaffold(); err != nil {
		return nil, err
	}

	restCfg := rest.Config{
		Port:      port,
		BodyLimit: "250K",
	}

	e := echo.New()
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))

	cacher := connectRedis(lg)

	return rest.NewBackOfficeAPI(e, restCfg, lg, ef, factory, cacher), nil
}

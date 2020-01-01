package rest

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/queue"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type API struct {
	e   *echo.Echo
	cfg Config
}

func New(cfg Config, q queue.MQ, mr model.MicroserviceRepository) *API {
	e := echo.New()

	e.Use(middleware.BodyLimit("250K"))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	h := handlers{
		logger:        e.Logger,
		microservices: mr,
	}

	e.POST("/events", h.CreateEvent)
	e.GET("/api/v1/microservices", h.SelectMicroservices)
	e.POST("/api/v1/microservices", h.CreateMicroservice)

	return &API{
		e:   e,
		cfg: cfg,
	}
}

func (a *API) Start() {
	go func() {
		if err := a.e.Start(a.cfg.Port); err != nil {
			a.e.Logger.Info("shutting down the server")
		}
	}()

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := a.e.Shutdown(ctx); err != nil {
		a.e.Logger.Fatal(err)
	}
}

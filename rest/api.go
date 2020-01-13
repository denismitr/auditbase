package rest

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type API struct {
	e   *echo.Echo
	cfg Config
}

func New(
	cfg Config,
	ef flow.EventFlow,
	mr model.MicroserviceRepository,
	er model.EventRepository,
) *API {
	e := echo.New()

	e.Use(middleware.BodyLimit(cfg.BodyLimit))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	mc := newMicroservicesController(e.Logger, mr)
	ec := newEventsController(e.Logger, er, ef)

	// Microservices
	e.GET("/api/v1/microservices", mc.SelectMicroservices)
	e.POST("/api/v1/microservices", mc.CreateMicroservice)
	e.PUT("/api/v1/microservices/:id", mc.UpdateMicroservice)
	e.GET("/api/v1/microservices/:id", mc.GetMicroservice)

	// Events
	e.POST("/api/v1/events", ec.CreateEvent)
	e.GET("/api/v1/events", ec.SelectEvents)
	e.GET("/api/v1/events/count", ec.Count)
	e.GET("/api/v1/events/queue", ec.Inspect)
	e.DELETE("/api/v1/events/:id", ec.DeleteEvent)
	e.GET("/api/v1/events/:id", ec.GetEvent)

	return &API{
		e:   e,
		cfg: cfg,
	}
}

// Start rest Server
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

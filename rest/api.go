package rest

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/clock"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

// API - rest API of auditbase
type API struct {
	e   *echo.Echo
	cfg Config
}

// New API
func New(
	cfg Config,
	logger logger.Logger,
	ef flow.EventFlow,
	microservices model.MicroserviceRepository,
	events model.EventRepository,
	entities model.EntityRepository,
) *API {
	e := echo.New()

	e.Use(middleware.BodyLimit(cfg.BodyLimit))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(hashRequestBody)

	uuid4 := uuid.NewUUID4Generator()

	microservicesController := newMicroservicesController(logger, uuid4, microservices)
	eventsController := newEventsController(logger, uuid4, clock.New(), events, ef)
	entitiesController := newEntitiesController(logger, uuid4, clock.New(), entities)

	// Microservices
	e.GET("/api/v1/microservices", microservicesController.index)
	e.POST("/api/v1/microservices", microservicesController.create)
	e.PUT("/api/v1/microservices/:id", microservicesController.update)
	e.GET("/api/v1/microservices/:id", microservicesController.show)

	// Events
	e.POST("/api/v1/events", eventsController.create)
	e.GET("/api/v1/events", eventsController.index)
	e.GET("/api/v1/events/count", eventsController.count)
	e.GET("/api/v1/events/queue", eventsController.inspect)
	e.DELETE("/api/v1/events/:id", eventsController.delete)
	e.GET("/api/v1/events/:id", eventsController.show)

	// Entities
	e.GET("/api/v1/entities", entitiesController.index)
	e.GET("/api/v1/entities/:id", entitiesController.show)

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

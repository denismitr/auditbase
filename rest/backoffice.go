package rest

import (
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/clock"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func NewBackOfficeAPI(
	e *echo.Echo,
	cfg Config,
	log logger.Logger,
	ef flow.EventFlow,
	microservices model.MicroserviceRepository,
	events model.EventRepository,
	entities model.EntityRepository,
) *API {
	e.Use(middleware.Logger())
	e.Use(middleware.BodyLimit(cfg.BodyLimit))
	e.Use(middleware.Recover())
	e.Use(hashRequestBody)

	uuid4 := uuid.NewUUID4Generator()

	microservicesController := newMicroservicesController(log, uuid4, microservices)
	eventsController := newEventsController(log, uuid4, clock.New(), events, ef)
	entitiesController := newEntitiesController(log, uuid4, clock.New(), entities)

	// Microservices
	e.GET("/api/v1/microservices", microservicesController.index)
	e.POST("/api/v1/microservices", microservicesController.create)
	e.PUT("/api/v1/microservices/:id", microservicesController.update)
	e.GET("/api/v1/microservices/:id", microservicesController.show)

	// Events
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
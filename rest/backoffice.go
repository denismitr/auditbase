package rest

import (
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/service"
	"github.com/denismitr/auditbase/utils/clock"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

type BackOfficeServices struct {
	Microservices service.MicroserviceService
	Actions service.ActionService
	Entities service.EntityService
}

func BackOfficeAPI(
	e *echo.Echo,
	cfg Config,
	log logger.Logger,
	ef flow.ActionFlow,
	services BackOfficeServices,
) *API {
	e.Use(middleware.Logger())
	e.Use(middleware.BodyLimit(cfg.BodyLimit))
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.DefaultCORSConfig))

	uuid4 := uuid.NewUUID4Generator()

	microservicesController := newMicroservicesController(log, uuid4, services.Microservices)
	eventsController := newActionsController(log, uuid4, clock.New(), services.Actions, ef)
	entitiesController := newEntitiesController(log, uuid4, clock.New(), services.Entities)

	// Microservices
	e.GET("/api/v1/microservices", microservicesController.index)
	e.POST("/api/v1/microservices", microservicesController.create)
	e.PUT("/api/v1/microservices/:id", microservicesController.update)
	e.GET("/api/v1/microservices/:id", microservicesController.show)

	// Events
	e.GET("/api/v1/actions", eventsController.index)
	e.GET("/api/v1/actions/count", eventsController.count)
	//e.GET("/api/v1/actions/queue", eventsController.inspect)
	//e.DELETE("/api/v1/actions/:id", eventsController.delete)
	e.GET("/api/v1/actions/:id", eventsController.show)

	// Entities
	e.GET("/api/v1/entitiesController", entitiesController.index)
	e.GET("/api/v1/entitiesController/:id", entitiesController.show)

	return &API{
		e:   e,
		cfg: cfg,
	}
}
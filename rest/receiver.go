package rest

import (
	"github.com/denismitr/auditbase/cache"
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/clock"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func NewReceiverAPI(
	e *echo.Echo,
	cfg Config,
	log logger.Logger,
	events model.EventRepository,
	ef flow.EventFlow,
	cacher cache.Cacher,
) *API {
	e.Use(middleware.BodyLimit(cfg.BodyLimit))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(hashRequestBody)

	uuid4 := uuid.NewUUID4Generator()

	eventsController := newEventsController(log, uuid4, clock.New(), events, ef, cacher)

	e.POST("/api/v1/events", eventsController.create)

	return &API{
		e:   e,
		cfg: cfg,
	}
}

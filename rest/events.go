package rest

import (
	"fmt"
	"github.com/denismitr/auditbase/cache"
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/clock"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/labstack/echo"
)

type eventsController struct {
	logger logger.Logger
	uuid4  uuid.UUID4Generator
	events model.EventRepository
	ef     flow.EventFlow
	clock  clock.Clock
	cacher cache.Cacher
}

func newEventsController(
	l logger.Logger,
	uuid4 uuid.UUID4Generator,
	clock clock.Clock,
	events model.EventRepository,
	ef flow.EventFlow,
	cacher cache.Cacher,
) *eventsController {
	return &eventsController{
		logger: l,
		clock:  clock,
		uuid4:  uuid4,
		events: events,
		ef:     ef,
		cacher: cacher,
	}
}

func (ec *eventsController) index(ctx echo.Context) error {
	q := ctx.Request().URL.Query()

	s := createSort(q)
	f := createFilter(q, []string{
		"actorServiceId",
		"targetServiceId",
		"eventName",
		"targetEntityId",
		"actorEntityId",
		"targetId",
		"propertyId",
		"actorId",
	})
	p := createPagination(q, 25)

	events, meta, err := ec.events.Select(f, s, p)
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(200, newEventsResponse(events, meta))
}

func (ec *eventsController) show(ctx echo.Context) error {
	ID := extractIDParamFrom(ctx)
	if errBag := ID.Validate(); errBag.NotEmpty() {
		return ctx.JSON(validationFailed(errBag.All()...))
	}

	event, err := ec.events.FindOneByID(ID)
	if err != nil {
		return ctx.JSON(notFound(err))
	}

	return ctx.JSON(200, newEventResponse(event))
}

func (ec *eventsController) count(ctx echo.Context) error {
	count, err := ec.events.Count()
	if err != nil {
		return ctx.JSON(notFound(err))
	}

	return ctx.JSON(200, newEventCountResponse(count))
}

func (ec *eventsController) inspect(ctx echo.Context) error {
	state, err := ec.ef.Inspect()
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	var status string

	if state.OK() {
		status = "OK"
	} else {
		status = state.Error()
	}

	i := inspectResource{
		ConnectionStatus: status,
		Messages:         state.Messages,
		Consumers:        state.Consumers,
	}

	return ctx.JSON(200, i.ToJSON())
}

func (ec *eventsController) delete(ctx echo.Context) error {
	ID := extractIDParamFrom(ctx)
	if errBag := ID.Validate(); errBag.NotEmpty() {
		return ctx.JSON(validationFailed(errBag.All()...))
	}

	err := ec.events.Delete(ID)
	if err != nil {
		return ctx.JSON(notFound(err))
	}

	return ctx.JSON(204, nil)
}

func hashKey(hash string) string {
	return fmt.Sprintf("hash_key_%s", hash)
}

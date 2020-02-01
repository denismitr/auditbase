package rest

import (
	"strconv"

	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type eventsController struct {
	logger utils.Logger
	uuid4  utils.UUID4Generatgor
	events model.EventRepository
	ef     flow.EventFlow
	clock  utils.Clock
}

func newEventsController(
	l utils.Logger,
	uuid4 utils.UUID4Generatgor,
	clock utils.Clock,
	events model.EventRepository,
	ef flow.EventFlow,
) *eventsController {
	return &eventsController{
		logger: l,
		clock:  clock,
		uuid4:  uuid4,
		events: events,
		ef:     ef,
	}
}

func (ec *eventsController) CreateEvent(ctx echo.Context) error {
	e := model.Event{}

	if err := ctx.Bind(&e); err != nil {
		return ctx.JSON(badRequest(errors.New("unparsable event payload")))
	}

	if e.ID == "" {
		e.ID = ec.uuid4.Generate()
	}

	// TODO: add validation, should not be empty
	if e.EmittedAt == 0 {
		e.EmittedAt = ec.clock.CurrentTimestamp()
	}

	e.RegisteredAt = ec.clock.CurrentTimestamp()

	v := model.NewValidator()

	errors := e.Validate(v)
	if !errors.IsEmpty() {
		return ctx.JSON(validationFailed(errors, "event object validation failed"))
	}

	if err := ec.ef.Send(e); err != nil {
		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(respondAccepted())
}

func (ec *eventsController) SelectEvents(ctx echo.Context) error {
	f := createEventFilterFromContext(ctx)

	events, err := ec.events.Select(f, model.Sort{}, model.Pagination{})
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(200, map[string]interface{}{
		"data": events,
	})
}

func (ec *eventsController) GetEvent(ctx echo.Context) error {
	ID := ctx.Param("id")
	event, err := ec.events.FindOneByID(ID)
	if err != nil {
		return ctx.JSON(notFound(err))
	}

	return ctx.JSON(200, map[string]interface{}{
		"data": event,
	})
}

func (ec *eventsController) Count(ctx echo.Context) error {
	count, err := ec.events.Count()
	if err != nil {
		return ctx.JSON(notFound(err))
	}

	return ctx.JSON(200, map[string]interface{}{
		"data": map[string]int{"count": count},
	})
}

func (ec *eventsController) Inspect(ctx echo.Context) error {
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

func (ec *eventsController) DeleteEvent(ctx echo.Context) error {
	ID := ctx.Param("id")
	err := ec.events.Delete(ID)
	if err != nil {
		return ctx.JSON(notFound(err))
	}

	return ctx.JSON(204, nil)
}

func createEventFilterFromContext(ctx echo.Context) model.EventFilter {
	var emittedAtGt int
	var emittedAtLt int

	if ctx.QueryParam("filter[emittedAt][gt]") != "" {
		emittedAtGt, _ = strconv.Atoi(ctx.QueryParam("filter[emittedAt][gt]"))
	}

	if ctx.QueryParam("filter[emittedAt][lt]") != "" {
		emittedAtLt, _ = strconv.Atoi(ctx.QueryParam("filter[emittedAt][lt]"))
	}

	return model.EventFilter{
		ActorTypeID:     ctx.QueryParam("filter[actorTypeId]"),
		ActorID:         ctx.QueryParam("filter[actorId]"),
		ActorServiceID:  ctx.QueryParam("filter[actorServiceId]"),
		TargetID:        ctx.QueryParam("filter[targetId]"),
		TargetTypeID:    ctx.QueryParam("filter[targetTypeId]"),
		TargetServiceID: ctx.QueryParam("filter[targetServiceId]"),
		EmittedAtGt:     int64(emittedAtGt),
		EmittedAtLt:     int64(emittedAtLt),
	}
}

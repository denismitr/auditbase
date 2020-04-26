package rest

import (
	"strconv"

	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/clock"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type eventsController struct {
	logger logger.Logger
	uuid4  uuid.UUID4Generator
	events model.EventRepository
	ef     flow.EventFlow
	clock  clock.Clock
}

func newEventsController(
	l logger.Logger,
	uuid4 uuid.UUID4Generator,
	clock clock.Clock,
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

func (ec *eventsController) create(ctx echo.Context) error {
	req := new(CreateEvent)

	if err := ctx.Bind(req); err != nil {
		err = errors.Wrap(err, "unparsable event payload")
		ec.logger.Error(err)
		return ctx.JSON(badRequest(err))
	}

	errorBag := req.Validate()
	if errorBag.NotEmpty() {
		return ctx.JSON(validationFailed(errorBag.All()...))
	}

	e := req.ToEvent()

	if e.ID == "" {
		e.ID = ec.uuid4.Generate()
	}

	if e.EmittedAt.IsZero() {
		e.EmittedAt = ec.clock.CurrentTime()
	}

	e.RegisteredAt = ec.clock.CurrentTime()
	e.Hash = ctx.Request().Header.Get("Body-Hash")

	if err := ec.ef.Send(e); err != nil {
		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(respondAccepted())
}

func (ec *eventsController) index(ctx echo.Context) error {
	f := createEventFilterFromContext(ctx)

	events, err := ec.events.Select(f, model.Sort{}, model.Pagination{PerPage: 100})
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(200, newEventsResponse(events))
}

func (ec *eventsController) show(ctx echo.Context) error {
	ID := extractIDParamFrom(ctx)
	if errors := ID.Validate(); errors.NotEmpty() {
		return ctx.JSON(validationFailed(errors.All()...))
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
	if errors := ID.Validate(); errors.NotEmpty() {
		return ctx.JSON(validationFailed(errors.All()...))
	}

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
		ActorEntityID:    ctx.QueryParam("filter[actorEntityId]"),
		ActorEntityName:  ctx.QueryParam("filter[actorEntityName]"),
		ActorID:          ctx.QueryParam("filter[actorId]"),
		ActorServiceID:   ctx.QueryParam("filter[actorServiceId]"),
		TargetID:         ctx.QueryParam("filter[targetId]"),
		TargetEntityID:   ctx.QueryParam("filter[targetEntityId]"),
		TargetEntityName: ctx.QueryParam("filter[targetEntityName]"),
		TargetServiceID:  ctx.QueryParam("filter[targetServiceId]"),
		EmittedAtGt:      int64(emittedAtGt),
		EmittedAtLt:      int64(emittedAtLt),
	}
}

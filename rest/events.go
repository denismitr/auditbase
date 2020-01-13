package rest

import (
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type eventsController struct {
	logger echo.Logger
	events model.EventRepository
	ef     flow.EventFlow
}

func newEventsController(
	l echo.Logger,
	events model.EventRepository,
	ef flow.EventFlow,
) *eventsController {
	return &eventsController{
		logger: l,
		events: events,
		ef:     ef,
	}
}

func (ec *eventsController) CreateEvent(ctx echo.Context) error {
	e := model.Event{}

	if err := ctx.Bind(&e); err != nil {
		return ctx.JSON(badRequest(errors.New("unparsable event payload")))
	}

	v := model.NewValidator()

	errors := e.Validate(v)
	if !errors.IsEmpty() {
		return ctx.JSON(validationFailed(errors, "event object validation failed"))
	}

	if err := ec.ef.Send(e); err != nil {
		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(202, map[string]string{
		"status": "Accepted",
	})
}

func (ec *eventsController) SelectEvents(ctx echo.Context) error {
	events, err := ec.events.SelectAll()
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
	messages, consumers, err := ec.ef.Inspect()
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	i := inspectResource{
		Messages:  messages,
		Consumers: consumers,
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

package rest

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/queue"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type eventsController struct {
	logger   echo.Logger
	events   model.EventRepository
	exchange queue.EventExchange
}

func (ec *eventsController) CreateEvent(ctx echo.Context) error {
	e := model.Event{}

	if err := ctx.Bind(&e); err != nil {
		return ctx.JSON(badRequest(errors.New("unparsable event payload")))
	}

	v := model.NewValidator()
	errors := e.Validate(v)
	if errors.HasErrors() {
		return ctx.JSON(validationFailed(errors, "event object validation failed"))
	}

	if err := ec.exchange.Publish(e); err != nil {
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

func (ec *eventsController) DeleteEvent(ctx echo.Context) error {
	ID := ctx.Param("id")
	err := ec.events.Delete(ID)
	if err != nil {
		return ctx.JSON(notFound(err))
	}

	return ctx.JSON(204, nil)
}

package rest

import (
	"time"

	"github.com/denismitr/auditbase/model"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type eventsController struct {
	logger echo.Logger
	events model.EventRepository
}

func (ec *eventsController) CreateEvent(ctx echo.Context) error {
	e := model.Event{}

	if err := ctx.Bind(&e); err != nil {
		return ctx.JSON(badRequest(errors.New("unparsable event payload")))
	}

	if e.ID == "" {
		e.ID = uuid4()
	}

	if e.EmittedAt == "" {
		e.EmittedAt = time.Now().String()
	}

	e.RegisteredAt = "9999-12-31 23:59:59"

	if err := ec.events.Create(e); err != nil {
		return ctx.JSON(badRequest(err))
	}

	return ctx.JSON(202, nil)
}

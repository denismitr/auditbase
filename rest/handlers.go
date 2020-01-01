package rest

import (
	"github.com/denismitr/auditbase/model"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type handlers struct {
	logger        echo.Logger
	microservices model.MicroserviceRepository
}

func (h *handlers) CreateEvent(ctx echo.Context) error {
	return nil
}

func (h *handlers) CreateMicroservice(ctx echo.Context) error {
	m := model.Microservice{}

	if err := ctx.Bind(&m); err != nil {
		return ctx.JSON(badRequest(errors.Wrap(err, "could not parse request payload")))
	}

	if m.ID == "" {
		m.ID = uuid4()
	}

	if err := h.microservices.Create(m); err != nil {
		return ctx.JSON(internalError(err))
	}

	savedMicroservice, err := h.microservices.GetOneByID(m.ID)
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	var r = make(map[string]model.Microservice)

	r["data"] = savedMicroservice

	return ctx.JSON(201, r)
}

func (h *handlers) SelectMicroservices(ctx echo.Context) error {
	ms, err := h.microservices.SelectAll()
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	var r = make(map[string][]model.Microservice)

	r["data"] = ms

	return ctx.JSON(200, r)
}

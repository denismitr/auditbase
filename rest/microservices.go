package rest

import (
	"github.com/denismitr/auditbase/model"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type microservicesController struct {
	logger        echo.Logger
	microservices model.MicroserviceRepository
}

func (mc *microservicesController) CreateMicroservice(ctx echo.Context) error {
	m := model.Microservice{}

	if err := ctx.Bind(&m); err != nil {
		return ctx.JSON(badRequest(errors.Wrap(err, "could not parse request payload")))
	}

	if m.ID == "" {
		m.ID = uuid4()
	}

	if err := mc.microservices.Create(m); err != nil {
		return ctx.JSON(internalError(err))
	}

	savedMicroservice, err := mc.microservices.GetOneByID(m.ID)
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	var r = make(map[string]model.Microservice)

	r["data"] = savedMicroservice

	return ctx.JSON(201, r)
}

func (mc *microservicesController) SelectMicroservices(ctx echo.Context) error {
	ms, err := mc.microservices.SelectAll()
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	var r = make(map[string][]model.Microservice) // TODO: refactor

	r["data"] = ms

	return ctx.JSON(200, r)
}

func (mc *microservicesController) UpdateMicroservice(ctx echo.Context) error {
	m := model.Microservice{}
	if err := ctx.Bind(&m); err != nil {
		return ctx.JSON(badRequest(errors.Wrap(err, "could not parse JSON payload")))
	}

	ID := ctx.Param("id")

	if err := mc.microservices.Update(ID, m); err != nil {
		return ctx.JSON(badRequest(err))
	}

	updatedM, err := mc.microservices.GetOneByID(ID)
	if err != nil {
		return ctx.JSON(badRequest(err))
	}

	var r = make(map[string]model.Microservice) // TODO: refactor
	r["data"] = updatedM

	return ctx.JSON(200, r)
}

func (mc *microservicesController) GetMicroservice(ctx echo.Context) error {
	ID := ctx.Param("id")
	if ID == "" {
		return ctx.JSON(badRequest(errors.New("ID is empty")))
	}

	m, err := mc.microservices.GetOneByID(ID)
	if err != nil {
		return ctx.JSON(badRequest(err)) // TODO: refactor to not found
	}

	var r = make(map[string]model.Microservice) // TODO: refactor
	r["data"] = m

	return ctx.JSON(200, r)
}

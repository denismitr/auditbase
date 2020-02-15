package rest

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type microservicesController struct {
	logger        utils.Logger
	uuid4         utils.UUID4Generatgor
	microservices model.MicroserviceRepository
}

func newMicroservicesController(
	l utils.Logger,
	uuid4 utils.UUID4Generatgor,
	m model.MicroserviceRepository,
) *microservicesController {
	return &microservicesController{
		logger:        l,
		uuid4:         uuid4,
		microservices: m,
	}
}

func (mc *microservicesController) CreateMicroservice(ctx echo.Context) error {
	m := model.Microservice{}

	if err := ctx.Bind(&m); err != nil {
		return ctx.JSON(badRequest(errors.Wrap(err, "could not parse request payload")))
	}

	if m.ID == "" {
		m.ID = mc.uuid4.Generate()
	}

	errors := m.Validate(model.NewValidator())
	if errors.NotEmpty() {
		return ctx.JSON(validationFailed(errors, "could not create a microservice"))
	}

	if err := mc.microservices.Create(m); err != nil {
		return ctx.JSON(internalError(err))
	}

	savedMicroservice, err := mc.microservices.FirstByID(model.ID(m.ID))
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

	return ctx.JSON(200, newResponse(ms))
}

func (mc *microservicesController) UpdateMicroservice(ctx echo.Context) error {
	m := model.Microservice{}
	if err := ctx.Bind(&m); err != nil {
		return ctx.JSON(badRequest(errors.Wrap(err, "could not parse JSON payload")))
	}

	if errors := m.Validate(model.NewValidator()); errors.NotEmpty() {
		return ctx.JSON(validationFailed(errors, "bad input data"))
	}

	ID := model.ID(ctx.Param("id"))
	if errors := ID.Validate(model.NewValidator()); errors.NotEmpty() {
		return ctx.JSON(validationFailed(errors, ":id is invalid"))
	}

	if err := mc.microservices.Update(ID, m); err != nil {
		return ctx.JSON(badRequest(err))
	}

	updatedM, err := mc.microservices.FirstByID(ID)
	if err != nil {
		return ctx.JSON(badRequest(err))
	}

	return ctx.JSON(200, newResponse(newMicroserviceResource(updatedM)))
}

func (mc *microservicesController) GetMicroservice(ctx echo.Context) error {
	ID := model.ID(ctx.Param("id"))
	v := model.NewValidator()

	if errors := ID.Validate(v); errors.NotEmpty() {
		return ctx.JSON(validationFailed(errors, ":id is incorrect"))
	}

	m, err := mc.microservices.FirstByID(ID)
	if err != nil {
		if err == model.ErrMicroserviceNotFound {
			return ctx.JSON(
				notFound(errors.Wrapf(err, "could not get microservice with ID %s from database", ID)))
		}

		return ctx.JSON(badRequest(err))
	}

	r := newMicroserviceResource(m)

	return ctx.JSON(200, newResponse(r))
}

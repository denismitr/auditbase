package rest

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type microservicesController struct {
	logger        logger.Logger
	uuid4         uuid.UUID4Generator
	microservices model.MicroserviceRepository
}

func newMicroservicesController(
	l logger.Logger,
	uuid4 uuid.UUID4Generator,
	m model.MicroserviceRepository,
) *microservicesController {
	return &microservicesController{
		logger:        l,
		uuid4:         uuid4,
		microservices: m,
	}
}

func (mc *microservicesController) create(ctx echo.Context) error {
	m := new(model.Microservice)

	if err := ctx.Bind(m); err != nil {
		return ctx.JSON(badRequest(errors.Wrap(err, "could not parse request payload")))
	}

	if m.ID == "" {
		m.ID = mc.uuid4.Generate()
	}

	errors := m.Validate()
	if errors.NotEmpty() {
		return ctx.JSON(validationFailed(errors.All()...))
	}

	savedMicroservice, err := mc.microservices.Create(m)
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(201, newMicroserviceResponse(savedMicroservice))
}

func (mc *microservicesController) index(ctx echo.Context) error {
	ms, err := mc.microservices.SelectAll()
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(200, newMicroservicesResponse(ms))
}

func (mc *microservicesController) update(ctx echo.Context) error {
	m := new(model.Microservice)
	if err := ctx.Bind(m); err != nil {
		return ctx.JSON(badRequest(errors.Wrap(err, "could not parse JSON payload")))
	}

	if errors := m.Validate(); errors.NotEmpty() {
		return ctx.JSON(validationFailed(errors.All()...))
	}

	ID := model.ID(ctx.Param("id"))
	if errors := ID.Validate(); errors.NotEmpty() {
		return ctx.JSON(validationFailed(errors.All()...))
	}

	if err := mc.microservices.Update(ID, m); err != nil {
		return ctx.JSON(badRequest(err))
	}

	return ctx.JSON(200, newMicroserviceResponse(m))
}

func (mc *microservicesController) show(ctx echo.Context) error {
	ID := extractIDParamFrom(ctx)

	if errors := ID.Validate(); errors.NotEmpty() {
		return ctx.JSON(validationFailed(errors.All()...))
	}

	m, err := mc.microservices.FirstByID(ID)
	if err != nil {
		if err == ErrMicroserviceNotFound {
			return ctx.JSON(
				notFound(errors.Wrapf(err, "could not get microservice with ID %s from database", ID)))
		}

		return ctx.JSON(badRequest(err))
	}

	return ctx.JSON(200, newMicroserviceResponse(m))
}

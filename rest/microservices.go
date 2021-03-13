package rest

import (
	"context"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/service"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"time"
)

type microservicesController struct {
	lg  logger.Logger
	uuid4   uuid.UUID4Generator
	microservices service.MicroserviceService
}

func newMicroservicesController(
	lg logger.Logger,
	uuid4 uuid.UUID4Generator,
	microservices service.MicroserviceService,
) *microservicesController {
	return &microservicesController{
		lg:        lg,
		uuid4:         uuid4,
		microservices: microservices,
	}
}

func (mc *microservicesController) create(rCtx echo.Context) error {
	m := new(model.Microservice)

	if err := rCtx.Bind(m); err != nil {
		return rCtx.JSON(badRequest(errors.Wrap(err, "could not parse request payload")))
	}

	if m.ID == "" {
		m.ID = model.ID(mc.uuid4.Generate())
	}

	errs := m.Validate()
	if errs.NotEmpty() {
		return rCtx.JSON(validationFailed(errs.All()...))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	savedMicroservice, err := mc.microservices.Create(ctx, m)
	if err != nil {
		return rCtx.JSON(internalError(err))
	}

	return rCtx.JSON(201, itemResource{
		Data: savedMicroservice,
	})
}

func (mc *microservicesController) index(rCtx echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	ms, err := mc.microservices.SelectAll(ctx, nil)
	if err != nil {
		return rCtx.JSON(internalError(err))
	}

	return rCtx.JSON(200, itemResource{
		Data: ms,
	})
}

func (mc *microservicesController) update(rCtx echo.Context) error {
	m := new(model.Microservice)
	if err := rCtx.Bind(m); err != nil {
		return rCtx.JSON(badRequest(errors.Wrap(err, "could not parse JSON update payload")))
	}

	if errs := m.Validate(); errs.NotEmpty() {
		return rCtx.JSON(validationFailed(errs.All()...))
	}

	ID := model.ID(rCtx.Param("id"))
	if errs := ID.Validate(); errs.NotEmpty() {
		return rCtx.JSON(validationFailed(errs.All()...))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	updated, err := mc.microservices.Update(ctx, ID, m);
	if err != nil {
		return rCtx.JSON(badRequest(err))
	}

	return rCtx.JSON(200, itemResource{
		Data: updated,
	})
}

func (mc *microservicesController) show(rCtx echo.Context) error {
	ID := extractIDParamFrom(rCtx)

	if errs := ID.Validate(); errs.NotEmpty() {
		return rCtx.JSON(validationFailed(errs.All()...))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	m, err := mc.microservices.FirstByID(ctx, ID, nil)
	if err != nil {
		if err == ErrMicroserviceNotFound {
			return rCtx.JSON(
				notFound(errors.Wrapf(err, "could not get microservice with ID %s from database", ID)))
		}

		return rCtx.JSON(badRequest(err))
	}

	return rCtx.JSON(200, itemResource{
		Data: m,
	})
}

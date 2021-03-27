package rest

import (
	"context"
	"github.com/denismitr/auditbase/internal/model"
	"github.com/denismitr/auditbase/internal/service"
	"github.com/denismitr/auditbase/internal/utils/logger"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

type microservicesController struct {
	lg            logger.Logger
	microservices service.MicroserviceService
}

func newMicroservicesController(
	lg logger.Logger,
	microservices service.MicroserviceService,
) *microservicesController {
	return &microservicesController{
		lg:        lg,
		microservices: microservices,
	}
}

func (mc *microservicesController) create(rCtx echo.Context) error {
	m := new(model.Microservice)

	if err := rCtx.Bind(m); err != nil {
		return rCtx.JSON(badRequest(errors.Wrap(err, "could not parse request payload")))
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

	numericID, err := strconv.Atoi(rCtx.Param("id"))
	if err != nil {
		return rCtx.JSON(badRequest(errors.New("ID is not numeric")))
	}

	if numericID <= 0 {
		return rCtx.JSON(badRequest(errors.New("ID must be a positive integer")))
	}

	ID := model.ID(numericID)

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
	ID, err := extractIDParamFrom(rCtx)
	if err != nil {
		return rCtx.JSON(badRequest(err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	m, err := mc.microservices.FirstByID(ctx, ID, nil)
	if err != nil {
		if err == ErrMicroserviceNotFound {
			return rCtx.JSON(
				notFound(errors.Wrapf(err, "could not get microservice with ID %d from database", ID)))
		}

		return rCtx.JSON(badRequest(err))
	}

	return rCtx.JSON(200, itemResource{
		Data: m,
	})
}

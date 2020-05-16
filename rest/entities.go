package rest

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/clock"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/denismitr/auditbase/utils/validator"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type entities struct {
	logger logger.Logger
	uuid4  uuid.UUID4Generator
	er     model.EntityRepository
	clock  clock.Clock
}

func newEntitiesController(
	l logger.Logger,
	uuid4 uuid.UUID4Generator,
	clock clock.Clock,
	er model.EntityRepository,
) *entities {
	return &entities{
		logger: l,
		uuid4:  uuid4,
		er:     er,
		clock:  clock,
	}
}

func (e *entities) index(ctx echo.Context) error {
	q := ctx.Request().URL.Query()

	s := createSort(q)
	f := createFilter(q, []string{"serviceId"})
	p := createPagination(q, 50)

	entities, err := e.er.Select(f, s, p)
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(200, newEntitiesResponse(entities))
}

func (e *entities) show(ctx echo.Context) error {
	ID := ctx.Param("id")
	if ID == "" {
		return ctx.JSON(badRequest(errors.New("ID is missing")))
	}

	entity, err := e.er.FirstByID(ID)
	if err != nil {
		e.logger.Error(err)
		return ctx.JSON(notFound(err))
	}

	return ctx.JSON(200, newEntityResponse(entity))
}

func (e *entities) properties(ctx echo.Context) error {
	ID := ctx.Param("id")
	if ID == "" || ! validator.IsUUID4(ID) {
		return ctx.JSON(badRequest(errors.New("ID is missing or invalid")))
	}

	entity, err := e.er.FirstByID(ID)
	if err != nil {
		e.logger.Error(err)
		return ctx.JSON(notFound(err))
	}

	stats, err := e.er.Properties(ID)
	if err != nil {
		e.logger.Error(err)
		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(200, newEntityWithPropertiesResponse(entity, stats))
}

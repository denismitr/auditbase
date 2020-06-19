package rest

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/denismitr/auditbase/utils/validator"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type properties struct {
	logger logger.Logger
	uuid4  uuid.UUID4Generator
	repo model.PropertyRepository
}

func  newPropertiesController(
	uuid4  uuid.UUID4Generator,
	logger logger.Logger,
	repo model.PropertyRepository,
) *properties {
	return &properties{
		uuid4: uuid4,
		logger: logger,
		repo: repo,
	}
}

func (p *properties) index(ctx echo.Context) error {
	q := ctx.Request().URL.Query()
	s := createSort(q)
	f := createFilter(q, []string{"entityId", "name"})
	pg := createPagination(q, 50)

	props, meta, err := p.repo.Select(f, s, pg)
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(200, newPropertiesResponse(props, meta))
}

func (c *properties) show(ctx echo.Context) error {
	ID := ctx.Param("id")
	if ! validator.IsUUID4(ID) {
		return ctx.JSON(validationFailed(ErrInvalidUUID4))
	}

	property, err := c.repo.FirstByID(ID)
	if err != nil {
		if err == model.ErrChangeNotFound {
			return ctx.JSON(
				notFound(errors.Wrapf(err, "could not get property with ID %s from storage", ID)))
		}

		return ctx.JSON(badRequest(err))
	}

	return ctx.JSON(200, newPropertyResponse(property))
}

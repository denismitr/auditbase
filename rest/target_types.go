package rest

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type targetTypes struct {
	logger utils.Logger
	uuid4  utils.UUID4Generatgor
	ttr    model.TargetTypeRepository
	clock  utils.Clock
}

func newTargetTypes(
	l utils.Logger,
	uuid4 utils.UUID4Generatgor,
	clock utils.Clock,
	ttr model.TargetTypeRepository,
) *targetTypes {
	return &targetTypes{
		logger: l,
		uuid4:  uuid4,
		ttr:    ttr,
		clock:  clock,
	}
}

func (tt *targetTypes) index(ctx echo.Context) error {
	actorTypes, err := tt.ttr.Select()
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(200, map[string]interface{}{
		"data": actorTypes,
	})
}

func (tt *targetTypes) show(ctx echo.Context) error {
	ID := ctx.Param("id")
	if ID == "" {
		return ctx.JSON(badRequest(errors.New("ID is missing")))
	}

	targetType, err := tt.ttr.FirstByID(ID)
	if err != nil {
		tt.logger.Error(err)
		return ctx.JSON(notFound(err))
	}

	return ctx.JSON(200, map[string]interface{}{
		"data": targetType,
	})
}

package rest

import (
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
)

type actorTypes struct {
	logger utils.Logger
	uuid4  utils.UUID4Generatgor
	atr    model.ActorTypeRepository
	clock  utils.Clock
}

func newActorTypes(
	l utils.Logger,
	uuid4 utils.UUID4Generatgor,
	clock utils.Clock,
	atr model.ActorTypeRepository,
) *actorTypes {
	return &actorTypes{
		logger: l,
		uuid4:  uuid4,
		atr:    atr,
		clock:  clock,
	}
}

func (ac *actorTypes) index(ctx echo.Context) error {
	actorTypes, err := ac.atr.Select()
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(200, map[string]interface{}{
		"data": actorTypes,
	})
}

func (ac *actorTypes) show(ctx echo.Context) error {
	ID := ctx.Param("id")
	if ID == "" {
		return ctx.JSON(badRequest(errors.New("ID is missing")))
	}

	actorType, err := ac.atr.FirstByID(ID)
	if err != nil {
		ac.logger.Error(err)
		return ctx.JSON(notFound(err))
	}

	return ctx.JSON(200, map[string]interface{}{
		"data": actorType,
	})
}

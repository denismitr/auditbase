package rest

import (
	"context"
	"fmt"
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/service"
	"github.com/denismitr/auditbase/utils/clock"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/labstack/echo"
	"time"
)

type actionsController struct {
	logger  logger.Logger
	uuid4   uuid.UUID4Generator
	actions service.ActionService
	ef      flow.ActionFlow
	clock   clock.Clock
}

func newActionsController(
	l logger.Logger,
	uuid4 uuid.UUID4Generator,
	clock clock.Clock,
	actions service.ActionService,
	ef flow.ActionFlow,
) *actionsController {
	return &actionsController{
		logger: l,
		clock:  clock,
		uuid4:  uuid4,
		actions: actions,
		ef:     ef,
	}
}

func (ec *actionsController) index(rCtx echo.Context) error {
	q := rCtx.Request().URL.Query()

	f := createFilter(q, []string{
		"name",
		"parentId",
	})

	c := createCursor(q, 25, []string{"name"})

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	actions, err := ec.actions.Select(ctx, c, f)
	if err != nil {
		return rCtx.JSON(internalError(err))
	}

	return rCtx.JSON(200, collectionResource{
		Data: actions.Items,
		Meta: actions.Meta,
	})
}

func (ec *actionsController) show(rCtx echo.Context) error {
	ID := extractIDParamFrom(rCtx)
	if errBag := ID.Validate(); errBag.NotEmpty() {
		return rCtx.JSON(validationFailed(errBag.All()...))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	action, err := ec.actions.FirstByID(ctx, ID)
	if err != nil {
		return rCtx.JSON(notFound(err))
	}

	return rCtx.JSON(200, itemResource{
		Data: action,
	})
}

func (ec *actionsController) count(rCtx echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	count, err := ec.actions.Count(ctx)
	if err != nil {
		return rCtx.JSON(notFound(err))
	}

	return rCtx.JSON(200, itemResource{
		Data: map[string]int{"count": count},
	})
}

//func (ec *actionsController) inspect(ctx echo.Context) error {
//	state, err := ec.af.Inspect()
//	if err != nil {
//		return ctx.JSON(internalError(err))
//	}
//
//	var status string
//
//	if state.OK() {
//		status = "OK"
//	} else {
//		status = state.Error()
//	}
//
//	i := inspectResource{
//		ConnectionStatus: status,
//		Messages:         state.Messages,
//		Consumers:        state.Consumers,
//	}
//
//	return ctx.JSON(200, i.ToJSON())
//}

//func (ec *actionsController) delete(ctx echo.Context) error {
//	ID := extractIDParamFrom(ctx)
//	if errBag := ID.Validate(); errBag.NotEmpty() {
//		return ctx.JSON(validationFailed(errBag.All()...))
//	}
//
//	err := ec.events.Delete(ID)
//	if err != nil {
//		return ctx.JSON(notFound(err))
//	}
//
//	return ctx.JSON(204, nil)
//}

func hashKey(hash string) string {
	return fmt.Sprintf("hash_key_%s", hash)
}

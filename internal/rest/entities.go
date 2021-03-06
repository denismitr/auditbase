package rest

import (
	"context"
	"github.com/denismitr/auditbase/internal/model"
	"github.com/denismitr/auditbase/internal/service"
	"github.com/denismitr/auditbase/internal/utils/clock"
	"github.com/denismitr/auditbase/internal/utils/logger"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

type entitiesController struct {
	logger   logger.Logger
	entities service.EntityService
	clock    clock.Clock
}

func newEntitiesController(
	l logger.Logger,
	clock clock.Clock,
	entities service.EntityService,
) *entitiesController {
	return &entitiesController{
		logger:   l,
		entities: entities,
		clock:    clock,
	}
}

func (e *entitiesController) index(rCtx echo.Context) error {
	q := rCtx.Request().URL.Query()
	f := createFilter(q, []string{"externalId", "entityTypeId"})
	c := createCursor(q, 50, []string{"externalId", "entityTypeId", "updatedAt", "createdAt"})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	entities, err := e.entities.Select(ctx, f, c)
	if err != nil {
		return rCtx.JSON(internalError(err))
	}

	return rCtx.JSON(200, collectionResource{
		Data: entities.Items,
		Meta: entities.Meta,
	})
}

func (e *entitiesController) show(rCtx echo.Context) error {
	ID := rCtx.Param("id")
	if ID == "" {
		return rCtx.JSON(badRequest(errors.New("ID is missing")))
	}

	numericID, err := strconv.Atoi(ID)
	if err != nil {
		return rCtx.JSON(badRequest(errors.New("ID is not numeric")))
	}

	// todo: numericID > 0

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	entity, err := e.entities.FirstByID(ctx, model.ID(numericID))
	if err != nil {
		e.logger.Error(err)
		return rCtx.JSON(notFound(err))
	}

	return rCtx.JSON(200, itemResource{
		Data: entity,
	})
}

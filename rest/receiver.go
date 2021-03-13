package rest

import (
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/model"
	"github.com/denismitr/auditbase/utils/clock"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
)

func NewReceiverAPI(
	e *echo.Echo,
	cfg Config,
	lg logger.Logger,
	ef flow.ActionFlow,
) *API {
	e.Use(middleware.BodyLimit(cfg.BodyLimit))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(hashRequestBody)

	uuid4 := uuid.NewUUID4Generator()

	receiverController := &receiverController{
		lg:    lg,
		uuid4: uuid4,
		clock: clock.New(),
		af:    ef,
	}

	e.POST("/api/v1/events", receiverController.create)

	return &API{
		e:   e,
		cfg: cfg,
	}
}

type receiverController struct {
	lg    logger.Logger
	uuid4 uuid.UUID4Generator
	clock clock.Clock
	af    flow.ActionFlow
}

func (rc *receiverController) create(ctx echo.Context) error {
	newAction := new(model.NewAction)

	if err := ctx.Bind(newAction); err != nil {
		err = errors.Wrap(err, "could not parse incoming action payload")
		rc.lg.Error(err)
		return ctx.JSON(badRequest(err))
	}

	errorBag := newAction.Validate()
	if errorBag.NotEmpty() {
		return ctx.JSON(validationFailed(errorBag.All()...))
	}

	newAction.Hash = ctx.Request().Header.Get("Body-Hash")

	// fixme: decide whether we actually need this check here
	//found, err := rc.cacher.Has(hashKey(e.Hash))
	//if err != nil {
	//	return ctx.JSON(internalError(err))
	//}
	//
	//if found {
	//	return ctx.JSON(conflict(ErrEventAlreadyReceived, "event already processed"))
	//}

	if newAction.ID == "" {
		newAction.ID = rc.uuid4.Generate()
	}

	newAction.RegisteredAt.Time = rc.clock.CurrentTime()

	//if err := rc.cacher.CreateKey(hashKey(e.Hash), 1 * time.Minute); err != nil {
	//	return ctx.JSON(internalError(err))
	//}

	if err := rc.af.Send(newAction); err != nil {
		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(202, itemResource{
		Status: "accepted",
		Data: map[string]string{"id": newAction.ID},
	})
}

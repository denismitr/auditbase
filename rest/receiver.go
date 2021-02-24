package rest

import (
	"github.com/denismitr/auditbase/cache"
	"github.com/denismitr/auditbase/flow"
	"github.com/denismitr/auditbase/utils/clock"
	"github.com/denismitr/auditbase/utils/logger"
	"github.com/denismitr/auditbase/utils/uuid"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
	"time"
)

func NewReceiverAPI(
	e *echo.Echo,
	cfg Config,
	lg logger.Logger,
	ef flow.ActionFlow,
	cacher cache.Cacher,
) *API {
	e.Use(middleware.BodyLimit(cfg.BodyLimit))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(hashRequestBody)

	uuid4 := uuid.NewUUID4Generator()

	receiverController := &receiverController{
		lg: lg,
		uuid4: uuid4,
		clock: clock.New(),
		ef: ef,
		cacher: cacher,
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
	clock   clock.Clock
	ef    flow.ActionFlow
	cacher     cache.Cacher
}

func (rc *receiverController) create(ctx echo.Context) error {
	req := new(CreateEvent)

	if err := ctx.Bind(req); err != nil {
		err = errors.Wrap(err, "unparsable event payload")
		rc.lg.Error(err)
		return ctx.JSON(badRequest(err))
	}

	errorBag := req.Validate()
	if errorBag.NotEmpty() {
		return ctx.JSON(validationFailed(errorBag.All()...))
	}

	e := req.ToEvent()
	e.Hash = ctx.Request().Header.Get("Body-Hash")

	found, err := rc.cacher.Has(hashKey(e.Hash))
	if err != nil {
		return ctx.JSON(internalError(err))
	}

	if found {
		return ctx.JSON(conflict(ErrEventAlreadyReceived, "event already processed"))
	}

	if e.ID == "" {
		e.ID = rc.uuid4.Generate()
	}

	e.RegisteredAt.Time = rc.clock.CurrentTime()

	if err := rc.cacher.CreateKey(hashKey(e.Hash), 1 * time.Minute); err != nil {
		return ctx.JSON(internalError(err))
	}

	if err := rc.ef.Send(e); err != nil {
		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(respondAccepted("events", e.ID))
}

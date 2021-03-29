package rest

import (
	"github.com/denismitr/auditbase/internal/receiver"
	"github.com/denismitr/auditbase/internal/utils/clock"
	"github.com/denismitr/auditbase/internal/utils/logger"
	"github.com/denismitr/auditbase/internal/utils/validator"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func NewReceiverAPI(
	e *echo.Echo,
	cfg Config,
	lg logger.Logger,
	rc *receiver.Receiver,
) *API {
	e.Use(middleware.BodyLimit(cfg.BodyLimit))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	receiverController := &receiverController{
		lg:    lg,
		clock: clock.New(),
		rc:    rc,
	}

	e.POST("/api/v1/actions", receiverController.create)
	e.PATCH("/api/v1/actions", receiverController.update)

	return &API{
		e:   e,
		cfg: cfg,
	}
}

type receiverController struct {
	lg    logger.Logger
	clock clock.Clock
	rc    *receiver.Receiver
}

func (rc *receiverController) create(ctx echo.Context) error {
	reg, err := rc.rc.ReceiveOneForCreate(ctx.Request().Body);
	if err != nil {
		if vErr, ok := err.(*validator.ValidationErrors); ok {
			return ctx.JSON(validationFailed(vErr.All()...))
		}

		switch err {
		case receiver.ErrActionAlreadyProcessed:
			return ctx.JSON(conflict(err, "action already processed"))
		case receiver.ErrInvalidInput:
			return ctx.JSON(badRequest(err))
		}

		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(202, itemResource{
		Status: "accepted",
		Data: map[string]string{
			"hash": reg.Hash,
			"uid":  reg.UID,
			"registeredAt": reg.RegisteredAt.String(),
		},
	})
}

func (rc *receiverController) update(ctx echo.Context) error {
	reg, err := rc.rc.ReceiveOneForUpdate(ctx.Request().Body);
	if err != nil {
		if vErr, ok := err.(*validator.ValidationErrors); ok {
			return ctx.JSON(validationFailed(vErr.All()...))
		}

		switch err {
		case receiver.ErrActionAlreadyProcessed:
			return ctx.JSON(conflict(err, "action already processed"))
		case receiver.ErrInvalidInput:
			return ctx.JSON(badRequest(err))
		}

		return ctx.JSON(internalError(err))
	}

	return ctx.JSON(202, itemResource{
		Status: "accepted",
		Data: map[string]string{
			"hash": reg.Hash,
			"uid":  reg.UID,
			"registeredAt": reg.RegisteredAt.String(),
		},
	})
}

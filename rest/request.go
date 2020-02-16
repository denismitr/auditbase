package rest

import (
	"github.com/denismitr/auditbase/model"
	"github.com/labstack/echo"
)

func extractIDParamFrom(ctx echo.Context) model.ID {
	id := ctx.Param("id")
	return model.ID(id)
}

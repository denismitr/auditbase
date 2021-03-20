package rest

import (
	"fmt"
	"strings"

	"github.com/denismitr/auditbase/model"
	"github.com/labstack/echo"
)

func extractIDParamFrom(ctx echo.Context) model.ID {
	id := ctx.Param("id")
	return model.ID(id)
}

func interfaceToStringPointer(value interface{}) *string {
	var out string

	switch raw := value.(type) {
	case string:
		out = raw
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		out = fmt.Sprintf("%d", raw)
	case float64, float32:
		out = strings.TrimRight(fmt.Sprintf("%0.4f", raw), "0")
	case bool:
		trueOrFalse := value.(bool)
		if !trueOrFalse {
			out = "0"
		} else {
			out = "1"
		}
	default:
		return nil
	}

	if out != "" {
		return &out
	}

	return nil
}

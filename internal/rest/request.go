package rest

import (
	"fmt"
	"github.com/pkg/errors"
	"strconv"
	"strings"

	"github.com/denismitr/auditbase/internal/model"
	"github.com/labstack/echo"
)

func extractIDParamFrom(ctx echo.Context) (model.ID, error) {
	id := ctx.Param("id")
	numericID, err := strconv.Atoi(id)
	if err != nil {
		return 0, errors.Wrapf(err, "invalid numeric ID value [%s]", id)
	}
	if numericID <= 0 {
		return 0, errors.Wrapf(err, "numeric ID must be positive, instead got [%s]", id)
	}

	return model.ID(numericID), nil
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

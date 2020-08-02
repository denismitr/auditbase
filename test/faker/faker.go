package faker

import (
	"github.com/denismitr/auditbase/utils/random"
	"github.com/denismitr/auditbase/utils/types"
)

func WrappedString(prefix string, suffix string, length int) string {
	r := random.String(length)

	if prefix != "" {
		r = prefix + "_" + r
	}

	if suffix != "" {
		r = r + "_" + suffix
	}

	return r
}

func NumericID(min, max int) string {
	return random.NumericString(min, max)
}

func ChangeValue(nullable bool) *string {
	formats := []string{"string", "integer", "float"}
	if nullable {
		formats = append(formats, "null")
	}

	format := formats[random.Int(0, len(formats) - 1)]

	switch format {
	case "integer":
		return types.PointerToString(NumericID(1, 10000))
	case "string":
		return types.PointerToString(random.String(random.Int(5, 100)))
	case "float":
		return types.PointerToString(NumericID(1, 100000)) // fixme
	default:
		return nil
	}
}



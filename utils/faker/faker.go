package faker

import (
	"github.com/denismitr/auditbase/utils/random"
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

func WrappedPointerToString(prefix string, suffix string, length int) *string {
	r := WrappedString(prefix, suffix, length)

	return &r
}

func NumericID(min, max int) string {
	return random.NumericString(min, max)
}

func NumericIDAsPointer(min, max int) *string {
	s := random.NumericString(min, max)
	return &s
}

func ChangeValue(nullable bool) interface{} {
	formats := []string{"string", "integer", "float"}
	if nullable {
		formats = append(formats, "null")
	}

	format := formats[random.Int(0, len(formats) - 1)]

	switch format {
	case "integer":
		return random.Int(0, 100000)
	case "string":
		return random.String(random.Int(5, 100))
	case "float":
		return NumericID(1, 100000) // fixme
	default:
		return nil
	}
}



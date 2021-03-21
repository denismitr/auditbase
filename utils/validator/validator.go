package validator

import (
	"regexp"
	"time"
)

const (
	integer string = `^[0-9]+$`
	float string = `[+-]?([0-9]*[.])[0-9]+`
)

var (
    rxInteger = regexp.MustCompile(integer)
    rxFloat = regexp.MustCompile(float)
)

func IsEmptyString(s string) bool {
	return s == ""
}

func IsEmptyTime(t time.Time) bool {
	return t.IsZero()
}

func StringLenBetween(s string, min, max int) bool {
	return len(s) > min && len(s) < max
}

func StringLenBetweenOrEq(s string, min, max int) bool {
	return len(s) >= min && len(s) <= max
}

func StringLenGt(s string, min int) bool {
	return len(s) > min
}

func StringLenGte(s string, min int) bool {
	return len(s) > min
}

func StringLenLt(s string , max int) bool {
	return len(s) < max
}

func StringLenLte(s string , max int) bool {
	return len(s) <= max
}

func IsInteger(s string) bool {
	return rxInteger.MatchString(s)
}

func IsFloat(s string) bool {
	return rxFloat.MatchString(s)
}

package validator

import (
	"regexp"
)

const (
	uuid4 string = "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
)

var (
	rxUUID4 = regexp.MustCompile(uuid4)
)

func IsEmptyString(s string) bool {
	return s == ""
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

func IsUUID4(s string) bool {
	return rxUUID4.MatchString(s)
}

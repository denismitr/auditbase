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

func StringLenGt(s string, max int) bool {
	if len(s) > max {
		return true
	}

	return false
}

func IsUUID4(s string) bool {
	return rxUUID4.MatchString(s)
}

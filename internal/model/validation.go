package model

import "regexp"

const (
	UUID4 string = "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$"
)

var (
	rxUUID4 = regexp.MustCompile(UUID4)
)

type ValidationErrors map[string][]string

func (e ValidationErrors) NotEmpty() bool {
	return len(e) > 0
}

func (e ValidationErrors) IsEmpty() bool {
	return len(e) == 0
}

type Validator struct {
	errors ValidationErrors
}

func (v *Validator) IsEmptyString(s string) bool {
	return s == ""
}

func (v *Validator) IsUUID4(s string) bool {
	return rxUUID4.MatchString(s)
}

func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

func (v *Validator) Add(key, message string) {
	v.errors[key] = append(v.errors[key], message)
}

func NewValidator() Validator {
	return Validator{errors: make(map[string][]string)}
}

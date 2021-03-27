package validator

import "fmt"

type ValidationErrors struct {
	errors map[string][]error
}

func (ve *ValidationErrors) NotEmpty() bool {
	return len(ve.errors) > 0
}

func (ve *ValidationErrors) IsEmpty() bool {
	return len(ve.errors) == 0
}

// All errors as flat slice
func (ve *ValidationErrors) All() []error {
	var errors []error
	for _, bag := range ve.errors {
		for i := range bag {
			errors = append(errors, bag[i])
		}
	}
	return errors
}

func (ve *ValidationErrors) Add(key string, err error) {
	ve.errors[key] = append(ve.errors[key], err)
}

func (ve *ValidationErrors) First() (string, error) {
	for key, bag := range ve.errors {
		for i := range bag {
			return key, bag[i]
		}
	}

	return "", nil
}

func (ve *ValidationErrors) Error() string {
	key, err := ve.First()
	if err == nil {
		return ""
	}

	return fmt.Sprintf("%s : %s", key, err.Error()) // TODO
}

func NewValidationError() *ValidationErrors {
	return &ValidationErrors{errors: make(map[string][]error)}
}

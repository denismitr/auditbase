package model

import (
	"encoding/json"
	"github.com/denismitr/auditbase/utils/errtype"
)

const ErrMissingActionID = errtype.StringError("action ID is empty")
const ErrNameIsRequired = errtype.StringError("name is required")
const ErrInvalidID = errtype.StringError("invalid ID value")
const ErrActorIDEmpty = errtype.StringError("actorEntityID must not be empty")
const ErrServiceNameInvalid = errtype.StringError("microservice name failed validation")
const ErrMicroserviceDescriptionTooLong = errtype.StringError("microservice description is too long")
const ErrMicroserviceNotFound = errtype.StringError("microservice not found")

const ErrActorEntityEmpty = errtype.StringError("actorEntity must not be empty")
const ErrTargetEntityEmpty = errtype.StringError("targetEntity must not be empty")
const ErrActorServiceEmpty = errtype.StringError("actorService must not be empty")
const ErrTargetServiceEmpty = errtype.StringError("targetService must not be empty")
const ErrEmittedAtEmpty = errtype.StringError("emittedAt must not be empty")

type ErrField struct {
	Name  string `json:"name"`
	Error string `json:"message"`
}

type ErrFields []ErrField
type ErrCode int

const (
	ErrCodeValidationFailed = iota
	ErrCodeNotFound
)

type AppError struct {
	Err    error     `json:"message"`
	Status int       `json:"status"`
	Fields ErrFields `json:"errors,omitempty"`
	Code   ErrCode   `json:"code"`
}

func NewValidationError(err error, fields ...ErrField) error {
	return &AppError{
		Err:    err,
		Status: 422,
		Fields: fields,
		Code:   ErrCodeValidationFailed,
	}
}

func NewNotFoundError(err error) error {
	return &AppError{
		Err:    err,
		Status: 404,
		Code:   ErrCodeNotFound,
	}
}

func (ae *AppError) Error() string {
	b, err := json.Marshal(ae)
	if err != nil {
		panic("how could error serialization fail? " + err.Error())
	}

	return string(b)
}

package rest

import (
	"net/http"

	"github.com/denismitr/auditbase/utils/errtype"
)

const ErrMissingEventID = errtype.StringError("event ID is empty")
const ErrInvalidUUID4 = errtype.StringError("not a valid UUID4")
const ErrActorIDEmpty = errtype.StringError("ActorID must not be empty")
const ErrTargetIDEmpty = errtype.StringError("TargetID must not be empty")
const ErrActorEntityEmpty = errtype.StringError("ActorEntity must not be empty")
const ErrTargetEntityEmpty = errtype.StringError("TargetEntity must not be empty")
const ErrActorServiceEmpty = errtype.StringError("ActorService must not be empty")
const ErrTargetServiceEmpty = errtype.StringError("TargetService must not be empty")
const ErrMicroserviceNameTooLong = errtype.StringError("microservice name is too long")
const ErrMicroserviceDescriptionTooLong = errtype.StringError("microservice description is too long")
const ErrMicroserviceNotFound = errtype.StringError("not found")
const ErrEventAlreadyReceived = errtype.StringError("event already received recently")
const ErrEmptyPropertyName = errtype.StringError("property name cannot be empty")

const msgBadRequest = "Bad request"
const msgInternalError = "Auditbase internal error"
const msgNotFound = "Entities not found"
const msgValidationFailed = "Validation failed"
const msgConflict = "entity you are trying to create already exists"

type errorResource struct {
	Title   string `json:"title"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

type errorResponse struct {
	Status int             `json:"status"`
	Errors []errorResource `json:"errors"`
}

func newErrorResourceWithDetails(code string, title string, details string) errorResource {
	return errorResource{
		Title:   title,
		Code:    code,
		Details: details,
	}
}

func newErrorResponse(status int, resources []errorResource) *errorResponse {
	return &errorResponse{
		Status: status,
		Errors: resources,
	}
}

func badRequest(errors ...error) (int, *errorResponse) {
	resources := make([]errorResource, len(errors))

	for i := range errors {
		resources[i] = newErrorResourceWithDetails("", msgBadRequest, errors[i].Error())
	}

	return http.StatusBadRequest, newErrorResponse(http.StatusBadRequest, resources)
}

func internalError(errors ...error) (int, *errorResponse) {
	resources := make([]errorResource, len(errors))

	for i := range errors {
		resources[i] = newErrorResourceWithDetails("", msgInternalError, errors[i].Error())
	}

	return http.StatusInternalServerError, newErrorResponse(http.StatusInternalServerError, resources)
}

func conflict(err error, msg string) (int, *errorResponse) {
	resources := make([]errorResource, 1)
	resources[0] = newErrorResourceWithDetails("EVENT_ALREADY_RECEIVED", msg, err.Error())

	return http.StatusConflict, newErrorResponse(http.StatusConflict, resources)
}

func notFound(errors ...error) (int, *errorResponse) {
	resources := make([]errorResource, len(errors))

	for i := range errors {
		resources[i] = newErrorResourceWithDetails("", msgNotFound, errors[i].Error())
	}

	return http.StatusNotFound, newErrorResponse(http.StatusNotFound, resources)
}

func validationFailed(errors ...error) (int, *errorResponse) {
	resources := make([]errorResource, len(errors))

	for i := range errors {
		resources[i] = newErrorResourceWithDetails("", msgValidationFailed, errors[i].Error())
	}

	return http.StatusUnprocessableEntity, newErrorResponse(http.StatusUnprocessableEntity, resources)
}

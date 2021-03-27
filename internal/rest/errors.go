package rest

import (
	"net/http"

	"github.com/denismitr/auditbase/internal/utils/errtype"
)

const ErrMicroserviceNotFound = errtype.StringError("not found")

const msgBadRequest = "Bad request"
const msgInternalError = "Auditbase internal error"
const msgNotFound = "Entities not found"
const msgValidationFailed = "Validation failed"

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

// fixme
func validationFailed(errors ...error) (int, *errorResponse) {
	resources := make([]errorResource, len(errors))

	for i := range errors {
		resources[i] = newErrorResourceWithDetails("", msgValidationFailed, errors[i].Error())
	}

	return http.StatusUnprocessableEntity, newErrorResponse(http.StatusUnprocessableEntity, resources)
}

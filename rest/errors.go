package rest

import (
	"net/http"

	"github.com/denismitr/auditbase/model"
)

type errorResponse struct {
	Title   string              `json:"title"`
	Code    int                 `json:"code"`
	Details string              `json:"details"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

func newErrorResponse(code int, title string) *errorResponse {
	return &errorResponse{
		Title: title,
		Code:  code,
	}
}

func newErrorResponseWithDetails(code int, title string, details string) *errorResponse {
	return &errorResponse{
		Title:   title,
		Code:    code,
		Details: details,
	}
}

func newErrorResponseWithDetailsAndErrors(
	code int,
	title string,
	details string,
	errors map[string][]string,
) *errorResponse {
	return &errorResponse{
		Title:   title,
		Code:    code,
		Details: details,
		Errors:  errors,
	}
}

func badRequest(err error) (int, *errorResponse) {
	return http.StatusBadRequest, newErrorResponseWithDetails(http.StatusBadRequest, "Bad request", err.Error())
}

func internalError(err error) (int, *errorResponse) {
	return http.StatusInternalServerError, newErrorResponseWithDetails(http.StatusInternalServerError, "Whoops", err.Error())
}

func notFound(err error) (int, *errorResponse) {
	return http.StatusNotFound, newErrorResponseWithDetails(http.StatusBadRequest, "Not found", err.Error())
}

func validationFailed(errors model.ValidationErrors, details string) (int, *errorResponse) {
	return http.StatusUnprocessableEntity, newErrorResponseWithDetailsAndErrors(
		http.StatusUnprocessableEntity,
		"Validation failed",
		details,
		errors,
	)
}

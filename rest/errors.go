package rest

import "net/http"

type errorResponse struct {
	Title   string `json:"title"`
	Code    int    `json:"code"`
	Details string `json:"details"`
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

func badRequest(err error) (int, *errorResponse) {
	return http.StatusBadRequest, newErrorResponseWithDetails(http.StatusBadRequest, "Bad request", err.Error())
}

func internalError(err error) (int, *errorResponse) {
	return http.StatusInternalServerError, newErrorResponseWithDetails(http.StatusInternalServerError, "Whoops", err.Error())
}

func notFound(err error) (int, *errorResponse) {
	return http.StatusNotFound, newErrorResponseWithDetails(http.StatusBadRequest, "Not found", err.Error())
}

func validationFailed(details string) *errorResponse {
	return newErrorResponseWithDetails(http.StatusUnprocessableEntity, "Validation failed", details)
}

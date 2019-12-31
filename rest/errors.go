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

func badRequest(details string) *errorResponse {
	return newErrorResponseWithDetails(http.StatusBadRequest, "Bad request", details)
}

func validationFailed(details string) *errorResponse {
	return newErrorResponseWithDetails(http.StatusUnprocessableEntity, "Validation failed", details)
}

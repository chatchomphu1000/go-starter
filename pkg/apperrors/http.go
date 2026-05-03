package apperrors

import (
	"errors"
	"net/http"
)

// ErrorResponse is the JSON body returned on error.
type ErrorResponse struct {
	Code      string   `json:"code"`
	Message   string   `json:"message"`
	Details   []string `json:"details,omitempty"`
	RequestID string   `json:"request_id,omitempty"`
}

// codeToHTTPStatus maps AppError codes to HTTP status codes.
var codeToHTTPStatus = map[string]int{
	CodeBadRequest:         http.StatusBadRequest,
	CodeUnauthorized:       http.StatusUnauthorized,
	CodeForbidden:          http.StatusForbidden,
	CodeNotFound:           http.StatusNotFound,
	CodeConflict:           http.StatusConflict,
	CodeUnprocessable:      http.StatusUnprocessableEntity,
	CodeTooManyRequests:    http.StatusTooManyRequests,
	CodeInternal:           http.StatusInternalServerError,
	CodeUnavailable:        http.StatusServiceUnavailable,
	CodeValidationFailed:   http.StatusBadRequest,
	CodeInvalidToken:       http.StatusUnauthorized,
	CodeWeakPassword:       http.StatusBadRequest,
	CodeUserInactive:       http.StatusForbidden,
	CodeNotifierFailed:     http.StatusInternalServerError,
	CodeEmailExists:        http.StatusConflict,
	CodeInvalidEmail:       http.StatusBadRequest,
	CodeInvalidRole:        http.StatusBadRequest,
	CodeInvalidCredentials: http.StatusUnauthorized,
}

// HTTPStatus extracts status code and ErrorResponse from an error.
// If the error is an AppError, the code is mapped; otherwise 500.
func HTTPStatus(err error, isProduction bool) (int, ErrorResponse) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		status := http.StatusInternalServerError
		if appErr.HTTPCode != 0 {
			status = appErr.HTTPCode
		} else if s, ok := codeToHTTPStatus[appErr.Code]; ok {
			status = s
		}

		msg := appErr.Message
		if isProduction && status >= 500 {
			msg = "internal server error"
		}

		return status, ErrorResponse{
			Code:    appErr.Code,
			Message: msg,
			Details: appErr.Details,
		}
	}

	msg := "internal server error"
	return http.StatusInternalServerError, ErrorResponse{
		Code:    CodeInternal,
		Message: msg,
	}
}

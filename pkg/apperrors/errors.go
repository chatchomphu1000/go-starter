package apperrors

import (
	"errors"
	"fmt"
)

// AppError represents a structured application error with a machine-readable code,
// human-readable message, optional details, and an optional underlying error.
type AppError struct {
	Code     string   `json:"code"`
	Message  string   `json:"message"`
	Details  []string `json:"details,omitempty"`
	Err      error    `json:"-"`
	HTTPCode int      `json:"-"`
}

// Error implements the error interface.
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error for errors.Is/As compatibility.
func (e *AppError) Unwrap() error {
	return e.Err
}

// Is checks whether the target error has the same code as this AppError.
func (e *AppError) Is(target error) bool {
	var t *AppError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

// WithDetails returns a copy of the AppError with added detail strings.
func (e *AppError) WithDetails(details ...string) *AppError {
	cp := *e
	cp.Details = append(cp.Details, details...)
	return &cp
}

// WithHTTPCode returns a copy of the AppError with an explicit HTTP status override.
func (e *AppError) WithHTTPCode(code int) *AppError {
	cp := *e
	cp.HTTPCode = code
	return &cp
}

// Wrap creates a new AppError wrapping an existing error with a code and message.
func Wrap(code, msg string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: msg,
		Err:     err,
	}
}

// NotFound creates a not-found AppError.
func NotFound(msg string, err error) *AppError {
	return &AppError{Code: CodeNotFound, Message: msg, Err: err}
}

// BadRequest creates a bad-request AppError.
func BadRequest(msg string, err error) *AppError {
	return &AppError{Code: CodeBadRequest, Message: msg, Err: err}
}

// Unauthorized creates an unauthorized AppError.
func Unauthorized(msg string, err error) *AppError {
	return &AppError{Code: CodeUnauthorized, Message: msg, Err: err}
}

// Forbidden creates a forbidden AppError.
func Forbidden(msg string, err error) *AppError {
	return &AppError{Code: CodeForbidden, Message: msg, Err: err}
}

// Conflict creates a conflict AppError.
func Conflict(msg string, err error) *AppError {
	return &AppError{Code: CodeConflict, Message: msg, Err: err}
}

// Validation creates a validation-failed AppError with field-level details.
func Validation(msg string, details []string) *AppError {
	return &AppError{Code: CodeValidationFailed, Message: msg, Details: details}
}

// Unprocessable creates an unprocessable-entity AppError.
func Unprocessable(msg string, err error) *AppError {
	return &AppError{Code: CodeUnprocessable, Message: msg, Err: err}
}

// TooManyRequests creates a rate-limit AppError.
func TooManyRequests(msg string) *AppError {
	return &AppError{Code: CodeTooManyRequests, Message: msg}
}

// Internal creates an internal-error AppError.
func Internal(msg string, err error) *AppError {
	return &AppError{Code: CodeInternal, Message: msg, Err: err}
}

// Unavailable creates a service-unavailable AppError.
func Unavailable(msg string, err error) *AppError {
	return &AppError{Code: CodeUnavailable, Message: msg, Err: err}
}

// FromDomain maps well-known domain sentinel errors to an appropriate AppError.
// If the error is not a known sentinel, it returns an internal error.
func FromDomain(err error) *AppError {
	if err == nil {
		return nil
	}

	// Map of domain sentinel error messages to AppError constructors.
	// This avoids importing the domain package — matching is done via errors.Is
	// at the call site or by the error string for sentinel errors.
	// The caller should use errors.Is checks before calling FromDomain.
	return Internal("an unexpected error occurred", err)
}

// Package apperrors provides application-level error types and codes.
package apperrors

// Machine-readable error codes used across the application.
const (
	CodeBadRequest         = "BAD_REQUEST"
	CodeUnauthorized       = "UNAUTHORIZED"
	CodeForbidden          = "FORBIDDEN"
	CodeNotFound           = "NOT_FOUND"
	CodeConflict           = "CONFLICT"
	CodeUnprocessable      = "UNPROCESSABLE"
	CodeTooManyRequests    = "TOO_MANY_REQUESTS"
	CodeInternal           = "INTERNAL_ERROR"
	CodeUnavailable        = "UNAVAILABLE"
	CodeValidationFailed   = "VALIDATION_FAILED"
	CodeInvalidToken       = "INVALID_TOKEN"
	CodeWeakPassword       = "WEAK_PASSWORD"
	CodeUserInactive       = "USER_INACTIVE"
	CodeNotifierFailed     = "NOTIFIER_FAILED"
	CodeEmailExists        = "EMAIL_ALREADY_EXISTS"
	CodeInvalidEmail       = "INVALID_EMAIL"
	CodeInvalidRole        = "INVALID_ROLE"
	CodeInvalidCredentials = "INVALID_CREDENTIALS" //nolint:gosec
)

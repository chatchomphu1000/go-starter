// Package dto contains HTTP request/response data transfer objects.
package dto

// ErrorResponse is the HTTP response body returned on errors.
type ErrorResponse struct {
	Code      string   `json:"code"`
	Message   string   `json:"message"`
	Details   []string `json:"details,omitempty"`
	RequestID string   `json:"request_id,omitempty"`
}

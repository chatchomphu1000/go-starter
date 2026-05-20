package apperrors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		isProduction bool
		wantStatus   int
		wantCode     string
		wantMsgMask  string // if non-empty, message must equal this
	}{
		{
			name:       "not_found",
			err:        NotFound("user not found", nil),
			wantStatus: http.StatusNotFound,
			wantCode:   CodeNotFound,
		},
		{
			name:       "bad_request",
			err:        BadRequest("bad input", nil),
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeBadRequest,
		},
		{
			name:       "unauthorized",
			err:        Unauthorized("not authenticated", nil),
			wantStatus: http.StatusUnauthorized,
			wantCode:   CodeUnauthorized,
		},
		{
			name:       "forbidden",
			err:        Forbidden("access denied", nil),
			wantStatus: http.StatusForbidden,
			wantCode:   CodeForbidden,
		},
		{
			name:       "conflict",
			err:        Conflict("already exists", nil),
			wantStatus: http.StatusConflict,
			wantCode:   CodeConflict,
		},
		{
			name:       "too_many_requests",
			err:        TooManyRequests("slow down"),
			wantStatus: http.StatusTooManyRequests,
			wantCode:   CodeTooManyRequests,
		},
		{
			name:       "internal",
			err:        Internal("oops", nil),
			wantStatus: http.StatusInternalServerError,
			wantCode:   CodeInternal,
		},
		{
			name:       "unavailable",
			err:        Unavailable("down", nil),
			wantStatus: http.StatusServiceUnavailable,
			wantCode:   CodeUnavailable,
		},
		{
			name:       "validation",
			err:        Validation("invalid", []string{"field: required"}),
			wantStatus: http.StatusBadRequest,
			wantCode:   CodeValidationFailed,
		},
		{
			name:         "production_masks_5xx_message",
			err:          Internal("secret internal detail", nil),
			isProduction: true,
			wantStatus:   http.StatusInternalServerError,
			wantCode:     CodeInternal,
			wantMsgMask:  "internal server error",
		},
		{
			name:         "production_does_not_mask_4xx",
			err:          NotFound("user not found", nil),
			isProduction: true,
			wantStatus:   http.StatusNotFound,
			wantCode:     CodeNotFound,
			wantMsgMask:  "user not found",
		},
		{
			name:       "explicit_http_code_override",
			err:        Internal("oops", nil).WithHTTPCode(http.StatusBadGateway),
			wantStatus: http.StatusBadGateway,
			wantCode:   CodeInternal,
		},
		{
			name:       "plain_error_returns_500",
			err:        errors.New("something random"),
			wantStatus: http.StatusInternalServerError,
			wantCode:   CodeInternal,
		},
		{
			name:       "wrapped_apperror_unwrapped_correctly",
			err:        NotFound("not found", errors.New("db miss")),
			wantStatus: http.StatusNotFound,
			wantCode:   CodeNotFound,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			status, resp := HTTPStatus(tc.err, tc.isProduction)
			assert.Equal(t, tc.wantStatus, status)
			assert.Equal(t, tc.wantCode, resp.Code)
			if tc.wantMsgMask != "" {
				assert.Equal(t, tc.wantMsgMask, resp.Message)
			}
		})
	}
}

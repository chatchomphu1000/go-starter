package apperrors

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppError_Error(t *testing.T) {
	t.Run("with_wrapped_error", func(t *testing.T) {
		cause := errors.New("db down")
		e := &AppError{Code: CodeNotFound, Message: "not found", Err: cause}
		assert.Contains(t, e.Error(), CodeNotFound)
		assert.Contains(t, e.Error(), "not found")
		assert.Contains(t, e.Error(), "db down")
	})

	t.Run("without_wrapped_error", func(t *testing.T) {
		e := &AppError{Code: CodeBadRequest, Message: "bad input"}
		assert.Contains(t, e.Error(), CodeBadRequest)
		assert.Contains(t, e.Error(), "bad input")
	})
}

func TestAppError_Unwrap(t *testing.T) {
	cause := errors.New("root cause")
	e := &AppError{Code: CodeInternal, Message: "oops", Err: cause}
	assert.Equal(t, cause, e.Unwrap())
	assert.True(t, errors.Is(e, cause))
}

func TestAppError_Is(t *testing.T) {
	e1 := &AppError{Code: CodeNotFound, Message: "msg1"}
	e2 := &AppError{Code: CodeNotFound, Message: "msg2"}
	e3 := &AppError{Code: CodeConflict, Message: "msg3"}

	assert.True(t, errors.Is(e1, e2), "same code should match")
	assert.False(t, errors.Is(e1, e3), "different code should not match")
}

func TestAppError_WithDetails(t *testing.T) {
	e := BadRequest("invalid input", nil)
	e2 := e.WithDetails("field: required", "email: invalid")
	assert.Equal(t, []string{"field: required", "email: invalid"}, e2.Details)
	assert.Empty(t, e.Details, "original should be unchanged")
}

func TestAppError_WithHTTPCode(t *testing.T) {
	e := Internal("oops", nil)
	e2 := e.WithHTTPCode(503)
	assert.Equal(t, 503, e2.HTTPCode)
	assert.Equal(t, 0, e.HTTPCode, "original should be unchanged")
}

func TestWrap(t *testing.T) {
	cause := errors.New("underlying")
	e := Wrap(CodeConflict, "conflict msg", cause)
	require.NotNil(t, e)
	assert.Equal(t, CodeConflict, e.Code)
	assert.Equal(t, "conflict msg", e.Message)
	assert.True(t, errors.Is(e, cause))
}

func TestConstructors(t *testing.T) {
	cause := errors.New("x")

	tests := []struct {
		name     string
		err      *AppError
		wantCode string
	}{
		{name: "NotFound", err: NotFound("msg", cause), wantCode: CodeNotFound},
		{name: "BadRequest", err: BadRequest("msg", cause), wantCode: CodeBadRequest},
		{name: "Unauthorized", err: Unauthorized("msg", cause), wantCode: CodeUnauthorized},
		{name: "Forbidden", err: Forbidden("msg", cause), wantCode: CodeForbidden},
		{name: "Conflict", err: Conflict("msg", cause), wantCode: CodeConflict},
		{name: "Unprocessable", err: Unprocessable("msg", cause), wantCode: CodeUnprocessable},
		{name: "TooManyRequests", err: TooManyRequests("msg"), wantCode: CodeTooManyRequests},
		{name: "Internal", err: Internal("msg", cause), wantCode: CodeInternal},
		{name: "Unavailable", err: Unavailable("msg", cause), wantCode: CodeUnavailable},
		{name: "Validation", err: Validation("msg", []string{"f1: required"}), wantCode: CodeValidationFailed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.NotNil(t, tc.err)
			assert.Equal(t, tc.wantCode, tc.err.Code)
		})
	}
}

func TestFromDomain(t *testing.T) {
	t.Run("nil_returns_nil", func(t *testing.T) {
		assert.Nil(t, FromDomain(nil))
	})

	t.Run("unknown_error_wraps_as_internal", func(t *testing.T) {
		e := FromDomain(errors.New("some error"))
		require.NotNil(t, e)
		assert.Equal(t, CodeInternal, e.Code)
	})
}

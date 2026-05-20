package handler_test

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	inboundhttp "github.com/chatchomphu1000/go-starter/internal/adapters/inbound/http"
	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

var fixedNow = time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)

func nopLogger() logger.Logger {
	l, _ := logger.NewLogger(logger.LoggerConfig{Level: "error", Format: "json"})
	return l
}

// makeEcho returns an Echo instance wired with validator and error handler.
func makeEcho() *echo.Echo {
	e := echo.New()
	e.Validator = inboundhttp.NewAppValidator()
	e.HTTPErrorHandler = inboundhttp.NewErrorHandler(nopLogger(), false)
	return e
}

// newRequest builds an httptest request/recorder pair and an Echo context.
func newRequest(t *testing.T, e *echo.Echo, method, path, body string) (echo.Context, *httptest.ResponseRecorder) {
	t.Helper()
	var reqBody *strings.Reader
	if body != "" {
		reqBody = strings.NewReader(body)
	} else {
		reqBody = strings.NewReader("")
	}
	req := httptest.NewRequest(method, path, reqBody)
	if body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	return c, rec
}

// call invokes the handler and pipes any error through the error handler.
func call(e *echo.Echo, c echo.Context, fn echo.HandlerFunc) {
	if err := fn(c); err != nil {
		e.HTTPErrorHandler(err, c)
	}
}

// validDomainUser returns a simple domain.User for use in mock returns.
func validDomainUser(t *testing.T) *domain.User {
	t.Helper()
	email, _ := domain.NewEmail("alice@example.com")
	u, _ := domain.NewUser("user-1", "Alice", email, "hashed", domain.RoleUser, fixedNow)
	return u
}

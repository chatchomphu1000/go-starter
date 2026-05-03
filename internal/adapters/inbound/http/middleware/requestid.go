// Package middleware provides HTTP middleware for the Echo server.
package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const headerXRequestID = "X-Request-ID"

// RequestID returns middleware that ensures every request has a unique X-Request-ID header.
// If the inbound request already contains a valid UUID, it is reused; otherwise a new one is generated.
func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqID := c.Request().Header.Get(headerXRequestID)
			if _, err := uuid.Parse(reqID); err != nil {
				reqID = uuid.New().String()
			}

			c.Set("request_id", reqID)
			c.Response().Header().Set(headerXRequestID, reqID)

			return next(c)
		}
	}
}

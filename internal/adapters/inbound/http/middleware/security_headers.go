package middleware

import (
	"strings"

	"github.com/labstack/echo/v4"
)

// SecurityHeaders returns middleware that sets common security headers.
// Swagger UI paths receive a relaxed CSP to allow its inline scripts and styles.
func SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			h := c.Response().Header()
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("X-Frame-Options", "DENY")
			h.Set("Referrer-Policy", "no-referrer")

			// Swagger UI requires inline scripts/styles and loads assets from 'self'.
			if strings.HasPrefix(c.Request().URL.Path, "/swagger/") {
				h.Set("Content-Security-Policy",
					"default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; frame-ancestors 'none'")
			} else {
				h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
			}

			if c.Request().TLS != nil {
				h.Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
			}

			return next(c)
		}
	}
}

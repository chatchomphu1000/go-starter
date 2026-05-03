package middleware

import (
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

// BodyLimit returns middleware that limits the request body size.
func BodyLimit(limit string) echo.MiddlewareFunc {
	return echomw.BodyLimit(limit)
}

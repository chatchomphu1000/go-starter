package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/pkg/apperrors"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// Recover returns middleware that recovers from panics and logs the stack trace.
func Recover(log logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) (returnErr error) {
			defer func() {
				if r := recover(); r != nil {
					stack := debug.Stack()
					log.Error("panic recovered",
						zap.Any("panic", r),
						zap.String("stack", string(stack)),
						zap.String("request_id", c.Response().Header().Get("X-Request-ID")),
					)

					var err error
					switch v := r.(type) {
					case error:
						err = v
					default:
						err = fmt.Errorf("%v", v)
					}

					returnErr = apperrors.Internal("internal server error", err).WithHTTPCode(http.StatusInternalServerError)
				}
			}()

			return next(c)
		}
	}
}

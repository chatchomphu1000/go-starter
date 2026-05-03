package middleware

import (
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// Logger returns middleware that logs HTTP requests using the given Logger.
// Paths in skipPaths are excluded from logging.
func Logger(log logger.Logger, skipPaths []string) echo.MiddlewareFunc {
	skip := make(map[string]struct{}, len(skipPaths))
	for _, p := range skipPaths {
		skip[p] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if _, ok := skip[c.Path()]; ok {
				return next(c)
			}

			start := time.Now()
			err := next(c)
			if err != nil {
				c.Error(err)
			}

			req := c.Request()
			res := c.Response()
			latency := time.Since(start)

			fields := []zap.Field{
				zap.String("request_id", res.Header().Get("X-Request-ID")),
				zap.String("method", req.Method),
				zap.String("uri", req.RequestURI),
				zap.Int("status", res.Status),
				zap.Int64("latency_ms", latency.Milliseconds()),
				zap.Int64("bytes_in", req.ContentLength),
				zap.Int64("bytes_out", res.Size),
				zap.String("ip", c.RealIP()),
				zap.String("user_agent", req.UserAgent()),
			}

			status := res.Status
			switch {
			case status >= 500:
				log.Error("request completed", fields...)
			case status >= 400:
				log.Warn("request completed", fields...)
			default:
				log.Info("request completed", fields...)
			}

			return nil
		}
	}
}

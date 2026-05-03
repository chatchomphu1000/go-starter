package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/chatchomphu1000/go-starter/pkg/apperrors"
)

// RateLimitConfig holds rate limiter configuration.
type RateLimitConfig struct {
	Enabled bool
	RPS     int
	Burst   int
}

type visitor struct {
	tokens   float64
	maxBurst float64
	rate     float64
	lastSeen time.Time
}

// RateLimit returns per-IP token-bucket rate limiting middleware.
func RateLimit(cfg RateLimitConfig) echo.MiddlewareFunc {
	if !cfg.Enabled {
		return func(next echo.HandlerFunc) echo.HandlerFunc {
			return next
		}
	}

	var (
		mu       sync.Mutex
		visitors = make(map[string]*visitor)
	)

	// Cleanup stale entries every minute.
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			mu.Lock()
			for ip, v := range visitors {
				if time.Since(v.lastSeen) > 3*time.Minute {
					delete(visitors, ip)
				}
			}
			mu.Unlock()
		}
	}()

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()
			now := time.Now()

			mu.Lock()
			v, ok := visitors[ip]
			if !ok {
				v = &visitor{
					tokens:   float64(cfg.Burst),
					maxBurst: float64(cfg.Burst),
					rate:     float64(cfg.RPS),
					lastSeen: now,
				}
				visitors[ip] = v
			}

			elapsed := now.Sub(v.lastSeen).Seconds()
			v.tokens += elapsed * v.rate
			if v.tokens > v.maxBurst {
				v.tokens = v.maxBurst
			}
			v.lastSeen = now

			if v.tokens < 1 {
				mu.Unlock()
				c.Response().Header().Set("Retry-After", "1")
				return apperrors.TooManyRequests("rate limit exceeded").WithHTTPCode(http.StatusTooManyRequests)
			}

			v.tokens--
			mu.Unlock()

			return next(c)
		}
	}
}

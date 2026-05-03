package http

import (
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/chatchomphu1000/go-starter/docs" // swag-generated docs
	"github.com/chatchomphu1000/go-starter/internal/adapters/inbound/http/handler"
	"github.com/chatchomphu1000/go-starter/internal/adapters/inbound/http/middleware"
	"github.com/chatchomphu1000/go-starter/internal/config"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/outbound"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// RouterConfig holds all dependencies required to set up routes.
type RouterConfig struct {
	Echo          *echo.Echo
	UserHandler   *handler.UserHandler
	HealthHandler *handler.HealthHandler
	TokenIssuer   outbound.TokenIssuer
	Config        *config.Config
	Logger        logger.Logger
}

// SetupRouter configures the Echo router with middleware and routes.
func SetupRouter(rc RouterConfig) {
	e := rc.Echo

	e.Validator = NewAppValidator()
	e.HTTPErrorHandler = NewErrorHandler(rc.Logger, rc.Config.IsProduction())

	skipPaths := []string{"/health", "/ready", "/version"}

	// Global middleware (order matters).
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger(rc.Logger, skipPaths))
	e.Use(middleware.Recover(rc.Logger))
	e.Use(middleware.SecurityHeaders())
	e.Use(middleware.CORS(middleware.CORSConfig{
		AllowOrigins:     rc.Config.CORS.AllowOrigins,
		AllowCredentials: rc.Config.CORS.AllowCredentials,
		MaxAge:           rc.Config.CORS.MaxAge,
	}))
	e.Use(middleware.BodyLimit(rc.Config.Server.BodyLimit))
	e.Use(middleware.RateLimit(middleware.RateLimitConfig{
		Enabled: rc.Config.RateLimit.Enabled,
		RPS:     rc.Config.RateLimit.RPS,
		Burst:   rc.Config.RateLimit.Burst,
	}))

	// Health endpoints (no auth, excluded from logger/rate-limiter via skipPaths).
	e.GET("/health", rc.HealthHandler.Liveness)
	e.GET("/ready", rc.HealthHandler.Readiness)
	e.GET("/version", rc.HealthHandler.Version)

	// API v1.
	v1 := e.Group("/api/v1")

	// Public routes.
	v1.POST("/users/register", rc.UserHandler.Register)
	v1.POST("/users/login", rc.UserHandler.Login)

	// Protected routes.
	protected := v1.Group("")
	protected.Use(middleware.Auth(rc.TokenIssuer))

	protected.GET("/users/:id", rc.UserHandler.GetByID)
	protected.GET("/users", rc.UserHandler.List)
	protected.PUT("/users/:id", rc.UserHandler.Update)
	protected.DELETE("/users/:id", rc.UserHandler.Delete)

	// Swagger (not in production unless explicitly enabled).
	if !rc.Config.IsProduction() || rc.Config.Swagger.Enabled {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}
}

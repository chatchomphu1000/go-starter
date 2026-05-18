package http

import (
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"

	_ "github.com/chatchomphu1000/go-starter/docs" // swag-generated docs
	"github.com/chatchomphu1000/go-starter/internal/adapters/inbound/http/handler"
	"github.com/chatchomphu1000/go-starter/internal/adapters/inbound/http/middleware"
	"github.com/chatchomphu1000/go-starter/internal/config"
	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/outbound"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// RouterConfig holds all dependencies required to set up routes.
type RouterConfig struct {
	Echo               *echo.Echo
	UserHandler        *handler.UserHandler
	AuthHandler        *handler.AuthHandler
	HealthHandler      *handler.HealthHandler
	RoomHandler        *handler.RoomHandler
	BookingHandler     *handler.BookingHandler
	PaymentHandler     *handler.PaymentHandler
	InvoiceHandler     *handler.InvoiceHandler
	MaintenanceHandler *handler.MaintenanceHandler
	NoticeHandler      *handler.NoticeHandler
	MessageHandler     *handler.MessageHandler
	ReportHandler      *handler.ReportHandler
	TokenIssuer        outbound.TokenIssuer
	Config             *config.Config
	Logger             logger.Logger
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

	// ── Public routes ─────────────────────────────────────────────────────────
	v1.POST("/auth/register", rc.AuthHandler.Register)
	v1.POST("/auth/login", rc.AuthHandler.Login)
	v1.POST("/auth/refresh", rc.AuthHandler.RefreshToken)

	// Public notices and rooms (read-only).
	v1.GET("/notices", rc.NoticeHandler.List)
	v1.GET("/notices/:id", rc.NoticeHandler.GetByID)
	v1.GET("/rooms", rc.RoomHandler.List)
	v1.GET("/rooms/:id", rc.RoomHandler.GetByID)

	// Webhook endpoint (no auth — secured by gateway signature verification).
	v1.POST("/payments/webhook/:gateway", rc.PaymentHandler.Webhook)

	// ── Protected routes (any authenticated user) ──────────────────────────
	auth := v1.Group("")
	auth.Use(middleware.Auth(rc.TokenIssuer))

	// User management.
	auth.GET("/users/:id", rc.UserHandler.GetByID)
	auth.PUT("/users/:id", rc.UserHandler.Update)
	auth.DELETE("/users/:id", rc.UserHandler.Delete)

	// Messages (all roles).
	auth.POST("/messages", rc.MessageHandler.Send)
	auth.GET("/messages/threads", rc.MessageHandler.ListThreads)
	auth.GET("/messages/threads/:threadId", rc.MessageHandler.ListMessages)
	auth.POST("/messages/threads/:threadId/read", rc.MessageHandler.MarkRead)

	// ── Owner routes ───────────────────────────────────────────────────────
	owner := v1.Group("")
	owner.Use(middleware.Auth(rc.TokenIssuer))
	owner.Use(middleware.RequireRole(domain.RoleOwner, domain.RoleAdmin))

	// User list (admin + owner).
	owner.GET("/users", rc.UserHandler.List)

	// Room management.
	owner.POST("/rooms", rc.RoomHandler.Create)
	owner.PUT("/rooms/:id", rc.RoomHandler.Update)
	owner.DELETE("/rooms/:id", rc.RoomHandler.Delete)

	// Booking lifecycle (owner side).
	owner.GET("/bookings", rc.BookingHandler.List)
	owner.GET("/bookings/:id", rc.BookingHandler.GetByID)
	owner.POST("/bookings/:id/approve", rc.BookingHandler.Approve)
	owner.POST("/bookings/:id/reject", rc.BookingHandler.Reject)
	owner.POST("/bookings/:id/activate", rc.BookingHandler.Activate)
	owner.POST("/bookings/:id/complete", rc.BookingHandler.Complete)

	// Invoice management.
	owner.POST("/invoices", rc.InvoiceHandler.Create)
	owner.GET("/invoices", rc.InvoiceHandler.List)
	owner.GET("/invoices/:id", rc.InvoiceHandler.GetByID)
	owner.POST("/invoices/:id/send", rc.InvoiceHandler.Send)
	owner.POST("/invoices/:id/pay", rc.InvoiceHandler.MarkPaid)
	owner.POST("/invoices/:id/cancel", rc.InvoiceHandler.Cancel)
	owner.GET("/invoices/:id/download", rc.InvoiceHandler.Download)

	// Payment management (owner).
	owner.GET("/payments", rc.PaymentHandler.List)
	owner.GET("/payments/:id", rc.PaymentHandler.GetByID)
	owner.POST("/payments/:id/refund", rc.PaymentHandler.Refund)

	// Maintenance (owner side).
	owner.GET("/maintenance", rc.MaintenanceHandler.List)
	owner.GET("/maintenance/:id", rc.MaintenanceHandler.GetByID)
	owner.POST("/maintenance/:id/start", rc.MaintenanceHandler.StartWork)
	owner.POST("/maintenance/:id/resolve", rc.MaintenanceHandler.Resolve)

	// Notice management.
	owner.POST("/notices", rc.NoticeHandler.Create)
	owner.PUT("/notices/:id", rc.NoticeHandler.Update)
	owner.DELETE("/notices/:id", rc.NoticeHandler.Delete)

	// Reports and dashboard.
	owner.GET("/owners/:id/dashboard", rc.ReportHandler.OwnerDashboard)
	owner.GET("/reports/income", rc.ReportHandler.IncomeExpense)

	// ── Tenant routes ──────────────────────────────────────────────────────
	tenant := v1.Group("")
	tenant.Use(middleware.Auth(rc.TokenIssuer))
	tenant.Use(middleware.RequireRole(domain.RoleTenant, domain.RoleUser, domain.RoleAdmin))

	// Booking (tenant creates and cancels).
	tenant.POST("/bookings", rc.BookingHandler.Create)
	tenant.POST("/bookings/:id/cancel", rc.BookingHandler.Cancel)

	// Invoice (tenant view and download).
	tenant.GET("/tenants/:id/invoices", rc.InvoiceHandler.List)
	tenant.GET("/tenants/:id/history", rc.ReportHandler.TenantStatement)

	// Payment (tenant creates payment).
	tenant.POST("/payments", rc.PaymentHandler.Create)

	// Maintenance (tenant opens and closes tickets).
	tenant.POST("/maintenance", rc.MaintenanceHandler.Create)
	tenant.POST("/maintenance/:id/close", rc.MaintenanceHandler.Close)

	// Swagger (not in production unless explicitly enabled).
	if !rc.Config.IsProduction() || rc.Config.Swagger.Enabled {
		e.GET("/swagger/*", echoSwagger.WrapHandler)
	}
}

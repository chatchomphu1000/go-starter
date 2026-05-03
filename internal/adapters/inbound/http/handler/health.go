// Package handler provides HTTP request handlers.
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// buildInfo holds version information injected at build time.
type buildInfo struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	BuildTime string `json:"build_time"`
}

// HealthHandler handles health and readiness checks.
type HealthHandler struct {
	mongoClient *mongo.Client
	log         logger.Logger
	build       buildInfo
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(mongoClient *mongo.Client, log logger.Logger, version, commit, buildTime string) *HealthHandler {
	return &HealthHandler{
		mongoClient: mongoClient,
		log:         log,
		build: buildInfo{
			Version:   version,
			Commit:    commit,
			BuildTime: buildTime,
		},
	}
}

// Liveness handles GET /health — returns 200 if the process is alive.
// @Summary      Liveness check
// @Description  Returns 200 OK if the server is running
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /health [get]
func (h *HealthHandler) Liveness(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// Readiness handles GET /ready — pings MongoDB with a 2s timeout.
// @Summary      Readiness check
// @Description  Returns 200 if MongoDB is reachable, 503 otherwise
// @Tags         health
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      503  {object}  map[string]string
// @Router       /ready [get]
func (h *HealthHandler) Readiness(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	if err := h.mongoClient.Ping(ctx, nil); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"status": "unavailable",
			"error":  "database unreachable",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// Version handles GET /version — returns build information.
// @Summary      Version info
// @Description  Returns version, commit, and build time
// @Tags         health
// @Produce      json
// @Success      200  {object}  handler.buildInfo
// @Router       /version [get]
func (h *HealthHandler) Version(c echo.Context) error {
	return c.JSON(http.StatusOK, h.build)
}

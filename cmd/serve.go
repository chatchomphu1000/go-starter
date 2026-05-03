package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	apphttp "github.com/chatchomphu1000/go-starter/internal/adapters/inbound/http"
	"github.com/chatchomphu1000/go-starter/internal/adapters/inbound/http/handler"
	"github.com/chatchomphu1000/go-starter/internal/adapters/outbound/clock"
	"github.com/chatchomphu1000/go-starter/internal/adapters/outbound/crypto"
	"github.com/chatchomphu1000/go-starter/internal/adapters/outbound/httpclient"
	"github.com/chatchomphu1000/go-starter/internal/adapters/outbound/idgen"
	"github.com/chatchomphu1000/go-starter/internal/adapters/outbound/jwtissuer"
	"github.com/chatchomphu1000/go-starter/internal/adapters/outbound/mongodb"
	"github.com/chatchomphu1000/go-starter/internal/core/services"
)

// version, commit, buildTime are set via ldflags.
var (
	version   = "dev"
	commit    = "none"
	buildTime = "unknown"
)

// SetBuildInfo sets the build information from main.go ldflags.
func SetBuildInfo(v, c, bt string) {
	version = v
	commit = c
	buildTime = bt
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		// Connect to MongoDB.
		mongoClient, err := mongodb.NewMongoClient(ctx, mongodb.MongoConfig{
			URI:      cfg.Mongo.URI,
			Database: cfg.Mongo.Database,
			MinPool:  cfg.Mongo.MinPool,
			MaxPool:  cfg.Mongo.MaxPool,
			Timeout:  cfg.Mongo.Timeout,
			AppName:  cfg.Mongo.AppName,
		}, log)
		if err != nil {
			return fmt.Errorf("failed to connect to MongoDB: %w", err)
		}

		// Run migrations if AUTO_MIGRATE is set.
		if os.Getenv("AUTO_MIGRATE") == "true" {
			runner, err := mongodb.NewMigrationRunner(cfg.Mongo.URI, cfg.Mongo.Database, "migrations", log)
			if err != nil {
				log.Warn("failed to create migration runner", zap.Error(err))
			} else {
				if err := runner.Up(); err != nil {
					log.Warn("migration failed", zap.Error(err))
				}
				_ = runner.Close()
			}
		}

		db := mongoClient.Database(cfg.Mongo.Database)

		// Build adapters.
		systemClock := clock.NewSystemClock()
		uuidGen := idgen.NewUUIDGenerator()
		hasher := crypto.NewBcryptHasher(12)
		tokenIssuer := jwtissuer.NewHS256Issuer(
			[]byte(cfg.JWT.Secret),
			cfg.JWT.TTL,
			cfg.JWT.Issuer,
			systemClock,
		)
		userRepo := mongodb.NewUserRepo(db)
		notifier := httpclient.NewNotifier(httpclient.NotifierConfig{
			BaseURL: cfg.Notifier.BaseURL,
			Timeout: cfg.Notifier.Timeout,
			Retry:   cfg.Notifier.Retry,
		}, log)

		// Build service.
		userService := services.NewUserService(userRepo, notifier, hasher, tokenIssuer, systemClock, uuidGen, log)

		// Build handlers.
		userHandler := handler.NewUserHandler(userService, log)
		healthHandler := handler.NewHealthHandler(mongoClient, log, version, commit, buildTime)

		// Build Echo.
		e := echo.New()
		e.HideBanner = true
		e.HidePort = true

		apphttp.SetupRouter(apphttp.RouterConfig{
			Echo:          e,
			UserHandler:   userHandler,
			HealthHandler: healthHandler,
			TokenIssuer:   tokenIssuer,
			Config:        cfg,
			Logger:        log,
		})

		// Start server in goroutine.
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		server := &http.Server{
			Addr:         addr,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		}
		e.Server = server

		go func() {
			log.Info("starting server", zap.String("addr", addr), zap.String("version", version))
			if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
				log.Error("server error", zap.Error(err))
			}
		}()

		// Wait for interrupt signal.
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Info("shutting down server...")

		// Graceful shutdown.
		shutdownCtx, cancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
		defer cancel()

		// 1. Shutdown HTTP server.
		if err := e.Shutdown(shutdownCtx); err != nil {
			log.Error("server shutdown error", zap.Error(err))
		} else {
			log.Info("server shutdown complete")
		}

		// 2. Close idle HTTP connections.
		notifier.CloseIdleConnections()

		// 3. Disconnect MongoDB.
		mongodb.Close(shutdownCtx, mongoClient, log)

		// 4. Sync logger.
		_ = log.Sync()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

package cmd

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

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
	"github.com/chatchomphu1000/go-starter/internal/worker"
	"github.com/chatchomphu1000/go-starter/internal/worker/jobs"
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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

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

		// ── Adapters ──────────────────────────────────────────────────────────
		systemClock := clock.NewSystemClock()
		uuidGen := idgen.NewUUIDGenerator()
		hasher := crypto.NewBcryptHasher(12)
		tokenIssuer := jwtissuer.NewHS256Issuer(
			[]byte(cfg.JWT.Secret),
			cfg.JWT.TTL,
			cfg.JWT.Issuer,
			systemClock,
		)
		notifier := httpclient.NewNotifier(httpclient.NotifierConfig{
			BaseURL: cfg.Notifier.BaseURL,
			Timeout: cfg.Notifier.Timeout,
			Retry:   cfg.Notifier.Retry,
		}, log)

		// ── Repositories ──────────────────────────────────────────────────────
		userRepo := mongodb.NewUserRepo(db)
		roomRepo := mongodb.NewRoomRepo(db)
		bookingRepo := mongodb.NewBookingRepo(db)
		paymentRepo := mongodb.NewPaymentRepo(db)
		invoiceRepo := mongodb.NewInvoiceRepo(db)
		maintenanceRepo := mongodb.NewMaintenanceRepo(db)
		noticeRepo := mongodb.NewNoticeRepo(db)
		messageRepo := mongodb.NewMessageRepo(db)
		activityLogRepo := mongodb.NewActivityLogRepo(db)
		_ = activityLogRepo // used by audit middleware in future extension

		// ── Services ──────────────────────────────────────────────────────────
		authService := services.NewAuthService(userRepo, notifier, hasher, tokenIssuer, systemClock, uuidGen, log)
		userService := services.NewUserService(userRepo, notifier, hasher, tokenIssuer, systemClock, uuidGen, log)
		roomService := services.NewRoomService(roomRepo, systemClock, uuidGen, log)
		bookingService := services.NewBookingService(bookingRepo, roomRepo, systemClock, uuidGen, log)
		paymentService := services.NewPaymentService(paymentRepo, invoiceRepo, systemClock, uuidGen, log)
		invoiceService := services.NewInvoiceService(invoiceRepo, userRepo, systemClock, uuidGen, log)
		maintenanceService := services.NewMaintenanceService(maintenanceRepo, systemClock, uuidGen, log)
		noticeService := services.NewNoticeService(noticeRepo, systemClock, uuidGen, log)
		messageService := services.NewMessageService(messageRepo, systemClock, uuidGen, log)
		reportService := services.NewReportService(
			roomRepo, bookingRepo, paymentRepo,
			invoiceRepo, maintenanceRepo, messageRepo, log,
		)

		// ── HTTP Handlers ─────────────────────────────────────────────────────
		authHandler := handler.NewAuthHandler(authService, log)
		userHandler := handler.NewUserHandler(userService, log)
		healthHandler := handler.NewHealthHandler(mongoClient, log, version, commit, buildTime)
		roomHandler := handler.NewRoomHandler(roomService, log)
		bookingHandler := handler.NewBookingHandler(bookingService, log)
		paymentHandler := handler.NewPaymentHandler(paymentService, log)
		invoiceHandler := handler.NewInvoiceHandler(invoiceService, log)
		maintenanceHandler := handler.NewMaintenanceHandler(maintenanceService, log)
		noticeHandler := handler.NewNoticeHandler(noticeService, log)
		messageHandler := handler.NewMessageHandler(messageService, log)
		reportHandler := handler.NewReportHandler(reportService, log)

		// ── Echo ──────────────────────────────────────────────────────────────
		e := echo.New()
		e.HideBanner = true
		e.HidePort = true

		apphttp.SetupRouter(apphttp.RouterConfig{
			Echo:               e,
			AuthHandler:        authHandler,
			UserHandler:        userHandler,
			HealthHandler:      healthHandler,
			RoomHandler:        roomHandler,
			BookingHandler:     bookingHandler,
			PaymentHandler:     paymentHandler,
			InvoiceHandler:     invoiceHandler,
			MaintenanceHandler: maintenanceHandler,
			NoticeHandler:      noticeHandler,
			MessageHandler:     messageHandler,
			ReportHandler:      reportHandler,
			TokenIssuer:        tokenIssuer,
			Config:             cfg,
			Logger:             log,
		})

		// ── Background Worker ─────────────────────────────────────────────────
		w := worker.New(3, 50, log)
		scheduler := worker.NewScheduler(w, log)

		scheduler.Register("invoice.overdue", 1*time.Hour,
			jobs.NewInvoiceOverdueJob(invoiceRepo, systemClock, log))
		scheduler.Register("rent.reminder", 24*time.Hour,
			jobs.NewRentReminderJob(invoiceRepo, userRepo, notifier, systemClock, log))

		// Start worker pool and scheduler in background goroutines.
		workerCtx, workerCancel := context.WithCancel(ctx)
		go w.Start(workerCtx)
		go scheduler.Run(workerCtx)

		// ── HTTP Server ───────────────────────────────────────────────────────
		addr := fmt.Sprintf(":%d", cfg.Server.Port)
		server := &http.Server{
			Addr:         addr,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  cfg.Server.IdleTimeout,
		}
		e.Server = server

		go func() {
			log.Info("starting server",
				zap.String("addr", addr),
				zap.String("version", version),
			)
			if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
				log.Error("server error", zap.Error(err))
			}
		}()

		// ── Graceful Shutdown ─────────────────────────────────────────────────
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		log.Info("shutting down server...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer shutdownCancel()

		// 1. Stop accepting new requests.
		if err := e.Shutdown(shutdownCtx); err != nil {
			log.Error("server shutdown error", zap.Error(err))
		} else {
			log.Info("server shutdown complete")
		}

		// 2. Stop background workers.
		workerCancel()

		// 3. Close idle HTTP connections.
		notifier.CloseIdleConnections()

		// 4. Disconnect MongoDB.
		mongodb.Close(shutdownCtx, mongoClient, log)

		// 5. Sync logger.
		_ = log.Sync()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

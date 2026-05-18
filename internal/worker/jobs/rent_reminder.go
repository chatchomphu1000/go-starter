package jobs

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/outbound"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// RentReminderJob sends payment reminders for sent invoices near their due date.
type RentReminderJob struct {
	invoiceRepo outbound.InvoiceRepository
	notifier    outbound.Notifier
	userRepo    outbound.UserRepository
	clock       outbound.Clock
	log         logger.Logger
}

// NewRentReminderJob creates a new RentReminderJob.
func NewRentReminderJob(
	invoiceRepo outbound.InvoiceRepository,
	userRepo outbound.UserRepository,
	notifier outbound.Notifier,
	clock outbound.Clock,
	log logger.Logger,
) *RentReminderJob {
	return &RentReminderJob{
		invoiceRepo: invoiceRepo,
		userRepo:    userRepo,
		notifier:    notifier,
		clock:       clock,
		log:         log,
	}
}

// Name returns the job identifier.
func (j *RentReminderJob) Name() string { return "rent.reminder" }

// Execute sends reminders for invoices due within the next 3 days.
func (j *RentReminderJob) Execute(ctx context.Context) error {
	now := j.clock.Now()
	reminderBefore := 3 * 24 * time.Hour

	sent := domain.InvoiceStatusSent
	invoices, _, err := j.invoiceRepo.FindAll(ctx, inbound.InvoiceFilter{
		Status: &sent,
		Page:   1,
		Limit:  100,
	})
	if err != nil {
		return fmt.Errorf("RentReminderJob: fetch invoices: %w", err)
	}

	reminded := 0
	for _, inv := range invoices {
		timeUntilDue := inv.DueDate.Sub(now)
		if timeUntilDue < 0 || timeUntilDue > reminderBefore {
			continue
		}

		tenant, err := j.userRepo.FindByID(ctx, inv.TenantID)
		if err != nil {
			j.log.Warn("reminder: tenant not found", zap.String("tenant_id", inv.TenantID), zap.Error(err))
			continue
		}

		bgCtx := context.WithoutCancel(ctx)
		if err := j.notifier.SendWelcomeEmail(bgCtx, tenant.Email, tenant.Name); err != nil {
			j.log.Warn("reminder: notification failed",
				zap.String("invoice_id", inv.ID),
				zap.String("tenant_id", inv.TenantID),
				zap.Error(err),
			)
			continue
		}
		reminded++
	}

	j.log.Info("rent reminder job complete",
		zap.Int("checked", len(invoices)),
		zap.Int("reminded", reminded),
	)
	return nil
}

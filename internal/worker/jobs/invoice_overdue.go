// Package jobs contains scheduled background job implementations.
package jobs

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/internal/core/ports/outbound"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

// InvoiceOverdueJob marks sent invoices past their due date as overdue.
type InvoiceOverdueJob struct {
	repo  outbound.InvoiceRepository
	clock outbound.Clock
	log   logger.Logger
}

// NewInvoiceOverdueJob creates a new InvoiceOverdueJob.
func NewInvoiceOverdueJob(repo outbound.InvoiceRepository, clock outbound.Clock, log logger.Logger) *InvoiceOverdueJob {
	return &InvoiceOverdueJob{repo: repo, clock: clock, log: log}
}

// Name returns the job identifier.
func (j *InvoiceOverdueJob) Name() string { return "invoice.overdue_check" }

// Execute finds and marks overdue invoices.
func (j *InvoiceOverdueJob) Execute(ctx context.Context) error {
	invoices, err := j.repo.FindOverdue(ctx)
	if err != nil {
		return fmt.Errorf("InvoiceOverdueJob: %w", err)
	}

	now := j.clock.Now()
	for _, inv := range invoices {
		inv.MarkOverdue(now)
		if err := j.repo.Update(ctx, inv); err != nil {
			j.log.Error("failed to mark invoice overdue",
				zap.String("invoice_id", inv.ID),
				zap.Error(err),
			)
		}
	}

	j.log.Info("overdue invoice check complete", zap.Int("marked", len(invoices)))
	return nil
}

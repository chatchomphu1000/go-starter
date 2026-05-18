package services

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/outbound"
	"github.com/chatchomphu1000/go-starter/pkg/apperrors"
	"github.com/chatchomphu1000/go-starter/pkg/logger"
)

type paymentService struct {
	repo        outbound.PaymentRepository
	invoiceRepo outbound.InvoiceRepository
	clock       outbound.Clock
	ids         outbound.IDGenerator
	log         logger.Logger
}

// NewPaymentService creates a new PaymentService.
func NewPaymentService(
	repo outbound.PaymentRepository,
	invoiceRepo outbound.InvoiceRepository,
	clock outbound.Clock,
	ids outbound.IDGenerator,
	log logger.Logger,
) inbound.PaymentService {
	return &paymentService{repo: repo, invoiceRepo: invoiceRepo, clock: clock, ids: ids, log: log}
}

// Create initiates a payment record.
func (s *paymentService) Create(ctx context.Context, in inbound.CreatePaymentInput) (*domain.Payment, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("paymentService.Create: %w", err)
	}

	p, err := domain.NewPayment(
		s.ids.New(), in.BookingID, in.InvoiceID,
		in.TenantID, in.OwnerID,
		in.Amount, in.Currency,
		domain.PaymentMethod(in.Method),
		in.Gateway, in.Description,
		s.clock.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("paymentService.Create: %w", apperrors.BadRequest(err.Error(), err))
	}

	if err := s.repo.Insert(ctx, p); err != nil {
		return nil, fmt.Errorf("paymentService.Create: %w", err)
	}

	s.log.Info("payment created", zap.String("payment_id", p.ID), zap.Float64("amount", in.Amount))
	return p, nil
}

// GetByID retrieves a payment by ID.
func (s *paymentService) GetByID(ctx context.Context, id string) (*domain.Payment, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("paymentService.GetByID: %w", err)
	}
	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("paymentService.GetByID: %w", err)
	}
	return p, nil
}

// List returns filtered payments.
func (s *paymentService) List(ctx context.Context, f inbound.PaymentFilter) ([]*domain.Payment, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, fmt.Errorf("paymentService.List: %w", err)
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	return s.repo.FindAll(ctx, f)
}

// HandleWebhook processes incoming webhook notifications from payment gateways.
func (s *paymentService) HandleWebhook(ctx context.Context, in inbound.WebhookInput) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("paymentService.HandleWebhook: %w", err)
	}

	p, err := s.repo.FindByID(ctx, in.PaymentID)
	if err != nil {
		return fmt.Errorf("paymentService.HandleWebhook: %w", err)
	}

	now := s.clock.Now()
	p.WebhookPayload = in.RawPayload

	switch in.Status {
	case "completed":
		if err := p.Complete(in.GatewayTxID, now); err != nil {
			// Already paid — idempotent, not an error
			s.log.Warn("webhook: payment already completed", zap.String("payment_id", p.ID))
			return nil
		}
		// Mark associated invoice as paid
		if p.InvoiceID != "" {
			if inv, invErr := s.invoiceRepo.FindByID(ctx, p.InvoiceID); invErr == nil {
				_ = inv.MarkPaid(now)
				_ = s.invoiceRepo.Update(ctx, inv)
			}
		}
	case "failed":
		p.Fail(now)
	case "refunded":
		_ = p.Refund(now)
	default:
		s.log.Warn("webhook: unknown status", zap.String("status", in.Status), zap.String("gateway", in.Gateway))
		return nil
	}

	if err := s.repo.Update(ctx, p); err != nil {
		return fmt.Errorf("paymentService.HandleWebhook: %w", err)
	}

	s.log.Info("webhook processed",
		zap.String("payment_id", p.ID),
		zap.String("gateway", in.Gateway),
		zap.String("status", in.Status),
	)
	return nil
}

// Refund refunds a completed payment — owner action.
func (s *paymentService) Refund(ctx context.Context, id, ownerID string) (*domain.Payment, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("paymentService.Refund: %w", err)
	}

	p, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("paymentService.Refund: %w", err)
	}
	if p.OwnerID != ownerID {
		return nil, fmt.Errorf("paymentService.Refund: %w", domain.ErrUnauthorizedAccess)
	}
	if err := p.Refund(s.clock.Now()); err != nil {
		return nil, fmt.Errorf("paymentService.Refund: %w", apperrors.BadRequest(err.Error(), err))
	}
	if err := s.repo.Update(ctx, p); err != nil {
		return nil, fmt.Errorf("paymentService.Refund: %w", err)
	}

	s.log.Info("payment refunded", zap.String("payment_id", id))
	return p, nil
}

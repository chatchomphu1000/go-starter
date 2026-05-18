package inbound

import (
	"context"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2 --name=PaymentService

// CreatePaymentInput holds data to initiate a payment.
type CreatePaymentInput struct {
	BookingID   string
	InvoiceID   string
	TenantID    string
	OwnerID     string
	Amount      float64
	Currency    string
	Method      string
	Gateway     string
	Description string
}

// WebhookInput holds the raw webhook payload from a payment gateway.
type WebhookInput struct {
	Gateway        string
	RawPayload     string
	GatewayTxID    string
	PaymentID      string
	Status         string // completed | failed | refunded
}

// PaymentFilter defines query parameters for listing payments.
type PaymentFilter struct {
	BookingID string
	InvoiceID string
	TenantID  string
	OwnerID   string
	Status    *domain.PaymentStatus
	Page      int
	Limit     int
	SortDesc  bool
}

// PaymentService defines the inbound port for payment processing.
type PaymentService interface {
	Create(ctx context.Context, in CreatePaymentInput) (*domain.Payment, error)
	GetByID(ctx context.Context, id string) (*domain.Payment, error)
	List(ctx context.Context, f PaymentFilter) ([]*domain.Payment, int64, error)
	HandleWebhook(ctx context.Context, in WebhookInput) error
	Refund(ctx context.Context, id string, ownerID string) (*domain.Payment, error)
}

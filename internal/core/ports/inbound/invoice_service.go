package inbound

import (
	"context"
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2 --name=InvoiceService

// InvoiceItemInput is a single line item for invoice creation.
type InvoiceItemInput struct {
	Description string
	Quantity    int
	UnitPrice   float64
}

// CreateInvoiceInput holds data required to generate an invoice.
type CreateInvoiceInput struct {
	BookingID string
	TenantID  string
	OwnerID   string
	Items     []InvoiceItemInput
	DueDate   time.Time
	Month     int
	Year      int
	Notes     string
}

// InvoiceFilter defines query parameters for listing invoices.
type InvoiceFilter struct {
	BookingID string
	TenantID  string
	OwnerID   string
	Status    *domain.InvoiceStatus
	Month     *int
	Year      *int
	Page      int
	Limit     int
	SortDesc  bool
}

// InvoiceService defines the inbound port for invoice management.
type InvoiceService interface {
	Create(ctx context.Context, in CreateInvoiceInput) (*domain.Invoice, error)
	GetByID(ctx context.Context, id string) (*domain.Invoice, error)
	List(ctx context.Context, f InvoiceFilter) ([]*domain.Invoice, int64, error)
	Send(ctx context.Context, id string, ownerID string) (*domain.Invoice, error)
	MarkPaid(ctx context.Context, id string, ownerID string) (*domain.Invoice, error)
	Cancel(ctx context.Context, id string, ownerID string) (*domain.Invoice, error)
	GeneratePDF(ctx context.Context, id string) ([]byte, string, error) // bytes, content-type, err
}

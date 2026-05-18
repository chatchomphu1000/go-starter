package outbound

import (
	"context"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
)

//go:generate go run github.com/vektra/mockery/v2 --name=InvoiceRepository

// InvoiceRepository defines the outbound port for invoice persistence.
type InvoiceRepository interface {
	Insert(ctx context.Context, inv *domain.Invoice) error
	FindByID(ctx context.Context, id string) (*domain.Invoice, error)
	FindAll(ctx context.Context, f inbound.InvoiceFilter) ([]*domain.Invoice, int64, error)
	Update(ctx context.Context, inv *domain.Invoice) error
	FindOverdue(ctx context.Context) ([]*domain.Invoice, error)
	CountByOwnerAndStatus(ctx context.Context, ownerID string, status domain.InvoiceStatus) (int64, error)
	SumByOwnerAndMonth(ctx context.Context, ownerID string, month, year int) (float64, int64, error)
}

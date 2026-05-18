package outbound

import (
	"context"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
)

//go:generate go run github.com/vektra/mockery/v2 --name=PaymentRepository

// PaymentRepository defines the outbound port for payment persistence.
type PaymentRepository interface {
	Insert(ctx context.Context, p *domain.Payment) error
	FindByID(ctx context.Context, id string) (*domain.Payment, error)
	FindAll(ctx context.Context, f inbound.PaymentFilter) ([]*domain.Payment, int64, error)
	Update(ctx context.Context, p *domain.Payment) error
	SumByOwnerAndMonth(ctx context.Context, ownerID string, month, year int) (float64, int64, error)
}

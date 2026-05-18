package outbound

import (
	"context"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
)

//go:generate go run github.com/vektra/mockery/v2 --name=MaintenanceRepository

// MaintenanceRepository defines the outbound port for maintenance ticket persistence.
type MaintenanceRepository interface {
	Insert(ctx context.Context, t *domain.MaintenanceTicket) error
	FindByID(ctx context.Context, id string) (*domain.MaintenanceTicket, error)
	FindAll(ctx context.Context, f inbound.TicketFilter) ([]*domain.MaintenanceTicket, int64, error)
	Update(ctx context.Context, t *domain.MaintenanceTicket) error
	CountOpenByOwner(ctx context.Context, ownerID string) (int64, error)
}

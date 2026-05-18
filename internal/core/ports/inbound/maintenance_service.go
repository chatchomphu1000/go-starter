package inbound

import (
	"context"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2 --name=MaintenanceService

// CreateTicketInput holds data to open a maintenance ticket.
type CreateTicketInput struct {
	RoomID      string
	TenantID    string
	OwnerID     string
	Title       string
	Description string
	Category    string
	Priority    string
	Photos      []string
}

// TicketFilter defines query parameters for listing tickets.
type TicketFilter struct {
	RoomID   string
	TenantID string
	OwnerID  string
	Status   *domain.TicketStatus
	Priority *domain.TicketPriority
	Page     int
	Limit    int
	SortDesc bool
}

// MaintenanceService defines the inbound port for maintenance ticket management.
type MaintenanceService interface {
	Create(ctx context.Context, in CreateTicketInput) (*domain.MaintenanceTicket, error)
	GetByID(ctx context.Context, id string) (*domain.MaintenanceTicket, error)
	List(ctx context.Context, f TicketFilter) ([]*domain.MaintenanceTicket, int64, error)
	StartWork(ctx context.Context, id string, ownerID string) (*domain.MaintenanceTicket, error)
	Resolve(ctx context.Context, id string, ownerID string, notes string) (*domain.MaintenanceTicket, error)
	Close(ctx context.Context, id string, tenantID string) (*domain.MaintenanceTicket, error)
}

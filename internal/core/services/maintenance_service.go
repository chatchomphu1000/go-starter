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

type maintenanceService struct {
	repo  outbound.MaintenanceRepository
	clock outbound.Clock
	ids   outbound.IDGenerator
	log   logger.Logger
}

// NewMaintenanceService creates a new MaintenanceService.
func NewMaintenanceService(
	repo outbound.MaintenanceRepository,
	clock outbound.Clock,
	ids outbound.IDGenerator,
	log logger.Logger,
) inbound.MaintenanceService {
	return &maintenanceService{repo: repo, clock: clock, ids: ids, log: log}
}

// Create opens a new maintenance ticket.
func (s *maintenanceService) Create(ctx context.Context, in inbound.CreateTicketInput) (*domain.MaintenanceTicket, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("maintenanceService.Create: %w", err)
	}

	priority := domain.TicketPriority(in.Priority)
	ticket, err := domain.NewMaintenanceTicket(
		s.ids.New(), in.RoomID, in.TenantID, in.OwnerID,
		in.Title, in.Description, in.Category,
		priority, in.Photos, s.clock.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("maintenanceService.Create: %w", apperrors.BadRequest(err.Error(), err))
	}

	if err := s.repo.Insert(ctx, ticket); err != nil {
		return nil, fmt.Errorf("maintenanceService.Create: %w", err)
	}

	s.log.Info("maintenance ticket created", zap.String("ticket_id", ticket.ID), zap.String("room_id", in.RoomID))
	return ticket, nil
}

// GetByID retrieves a ticket by ID.
func (s *maintenanceService) GetByID(ctx context.Context, id string) (*domain.MaintenanceTicket, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("maintenanceService.GetByID: %w", err)
	}
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("maintenanceService.GetByID: %w", err)
	}
	return t, nil
}

// List returns filtered maintenance tickets.
func (s *maintenanceService) List(ctx context.Context, f inbound.TicketFilter) ([]*domain.MaintenanceTicket, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, fmt.Errorf("maintenanceService.List: %w", err)
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	return s.repo.FindAll(ctx, f)
}

// StartWork owner begins working on a ticket.
func (s *maintenanceService) StartWork(ctx context.Context, id, ownerID string) (*domain.MaintenanceTicket, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("maintenanceService.StartWork: %w", err)
	}
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("maintenanceService.StartWork: %w", err)
	}
	if t.OwnerID != ownerID {
		return nil, fmt.Errorf("maintenanceService.StartWork: %w", domain.ErrUnauthorizedAccess)
	}
	if err := t.StartWork(s.clock.Now()); err != nil {
		return nil, fmt.Errorf("maintenanceService.StartWork: %w", apperrors.BadRequest(err.Error(), err))
	}
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("maintenanceService.StartWork: %w", err)
	}
	s.log.Info("ticket in progress", zap.String("ticket_id", id))
	return t, nil
}

// Resolve owner marks the ticket as resolved.
func (s *maintenanceService) Resolve(ctx context.Context, id, ownerID, notes string) (*domain.MaintenanceTicket, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("maintenanceService.Resolve: %w", err)
	}
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("maintenanceService.Resolve: %w", err)
	}
	if t.OwnerID != ownerID {
		return nil, fmt.Errorf("maintenanceService.Resolve: %w", domain.ErrUnauthorizedAccess)
	}
	if err := t.Resolve(notes, s.clock.Now()); err != nil {
		return nil, fmt.Errorf("maintenanceService.Resolve: %w", apperrors.BadRequest(err.Error(), err))
	}
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("maintenanceService.Resolve: %w", err)
	}
	s.log.Info("ticket resolved", zap.String("ticket_id", id))
	return t, nil
}

// Close tenant confirms the resolution and closes the ticket.
func (s *maintenanceService) Close(ctx context.Context, id, tenantID string) (*domain.MaintenanceTicket, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("maintenanceService.Close: %w", err)
	}
	t, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("maintenanceService.Close: %w", err)
	}
	if t.TenantID != tenantID {
		return nil, fmt.Errorf("maintenanceService.Close: %w", domain.ErrUnauthorizedAccess)
	}
	if err := t.Close(s.clock.Now()); err != nil {
		return nil, fmt.Errorf("maintenanceService.Close: %w", apperrors.BadRequest(err.Error(), err))
	}
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, fmt.Errorf("maintenanceService.Close: %w", err)
	}
	s.log.Info("ticket closed", zap.String("ticket_id", id))
	return t, nil
}

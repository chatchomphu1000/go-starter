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

type bookingService struct {
	repo     outbound.BookingRepository
	roomRepo outbound.RoomRepository
	clock    outbound.Clock
	ids      outbound.IDGenerator
	log      logger.Logger
}

// NewBookingService creates a new BookingService.
func NewBookingService(
	repo outbound.BookingRepository,
	roomRepo outbound.RoomRepository,
	clock outbound.Clock,
	ids outbound.IDGenerator,
	log logger.Logger,
) inbound.BookingService {
	return &bookingService{repo: repo, roomRepo: roomRepo, clock: clock, ids: ids, log: log}
}

// Create submits a new booking request from a tenant.
func (s *bookingService) Create(ctx context.Context, in inbound.CreateBookingInput) (*domain.Booking, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("bookingService.Create: %w", err)
	}

	room, err := s.roomRepo.FindByID(ctx, in.RoomID)
	if err != nil {
		return nil, fmt.Errorf("bookingService.Create: %w", err)
	}
	if !room.IsAvailable() {
		return nil, fmt.Errorf("bookingService.Create: %w", domain.ErrRoomUnavailable)
	}

	conflict, err := s.repo.HasActiveBooking(ctx, in.RoomID, in.StartDate, in.EndDate)
	if err != nil {
		return nil, fmt.Errorf("bookingService.Create: %w", err)
	}
	if conflict {
		return nil, fmt.Errorf("bookingService.Create: %w", domain.ErrBookingConflict)
	}

	booking, err := domain.NewBooking(
		s.ids.New(), in.RoomID, in.TenantID, room.OwnerID,
		in.StartDate, in.EndDate,
		room.RentPrice, room.Deposit,
		in.Notes, s.clock.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("bookingService.Create: %w", apperrors.BadRequest(err.Error(), err))
	}

	if err := s.repo.Insert(ctx, booking); err != nil {
		return nil, fmt.Errorf("bookingService.Create: %w", err)
	}

	s.log.Info("booking created", zap.String("booking_id", booking.ID), zap.String("room_id", in.RoomID))
	return booking, nil
}

// GetByID retrieves a booking by ID.
func (s *bookingService) GetByID(ctx context.Context, id string) (*domain.Booking, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("bookingService.GetByID: %w", err)
	}
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("bookingService.GetByID: %w", err)
	}
	return b, nil
}

// List returns filtered bookings.
func (s *bookingService) List(ctx context.Context, f inbound.BookingFilter) ([]*domain.Booking, int64, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, fmt.Errorf("bookingService.List: %w", err)
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 20
	}
	return s.repo.FindAll(ctx, f)
}

// Approve owner approves a pending booking.
func (s *bookingService) Approve(ctx context.Context, id, ownerID string) (*domain.Booking, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("bookingService.Approve: %w", err)
	}
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("bookingService.Approve: %w", err)
	}
	if b.OwnerID != ownerID {
		return nil, fmt.Errorf("bookingService.Approve: %w", domain.ErrUnauthorizedAccess)
	}
	if err := b.Approve(s.clock.Now()); err != nil {
		return nil, fmt.Errorf("bookingService.Approve: %w", apperrors.BadRequest(err.Error(), err))
	}
	if err := s.repo.Update(ctx, b); err != nil {
		return nil, fmt.Errorf("bookingService.Approve: %w", err)
	}
	s.log.Info("booking approved", zap.String("booking_id", id))
	return b, nil
}

// Reject owner rejects a pending booking.
func (s *bookingService) Reject(ctx context.Context, id, ownerID, reason string) (*domain.Booking, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("bookingService.Reject: %w", err)
	}
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("bookingService.Reject: %w", err)
	}
	if b.OwnerID != ownerID {
		return nil, fmt.Errorf("bookingService.Reject: %w", domain.ErrUnauthorizedAccess)
	}
	if err := b.Reject(reason, s.clock.Now()); err != nil {
		return nil, fmt.Errorf("bookingService.Reject: %w", apperrors.BadRequest(err.Error(), err))
	}
	if err := s.repo.Update(ctx, b); err != nil {
		return nil, fmt.Errorf("bookingService.Reject: %w", err)
	}
	s.log.Info("booking rejected", zap.String("booking_id", id))
	return b, nil
}

// Cancel tenant or owner cancels a booking.
func (s *bookingService) Cancel(ctx context.Context, id, requesterID string) (*domain.Booking, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("bookingService.Cancel: %w", err)
	}
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("bookingService.Cancel: %w", err)
	}
	if b.TenantID != requesterID && b.OwnerID != requesterID {
		return nil, fmt.Errorf("bookingService.Cancel: %w", domain.ErrUnauthorizedAccess)
	}
	if err := b.Cancel(s.clock.Now()); err != nil {
		return nil, fmt.Errorf("bookingService.Cancel: %w", apperrors.BadRequest(err.Error(), err))
	}
	if err := s.repo.Update(ctx, b); err != nil {
		return nil, fmt.Errorf("bookingService.Cancel: %w", err)
	}
	s.log.Info("booking cancelled", zap.String("booking_id", id))
	return b, nil
}

// Activate owner marks an approved booking as active (tenant moved in).
func (s *bookingService) Activate(ctx context.Context, id, ownerID string) (*domain.Booking, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("bookingService.Activate: %w", err)
	}
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("bookingService.Activate: %w", err)
	}
	if b.OwnerID != ownerID {
		return nil, fmt.Errorf("bookingService.Activate: %w", domain.ErrUnauthorizedAccess)
	}
	if err := b.Activate(s.clock.Now()); err != nil {
		return nil, fmt.Errorf("bookingService.Activate: %w", apperrors.BadRequest(err.Error(), err))
	}

	// Mark room as occupied.
	room, roomErr := s.roomRepo.FindByID(ctx, b.RoomID)
	if roomErr == nil {
		room.SetStatus(domain.RoomStatusOccupied)
		room.Touch(s.clock.Now())
		_ = s.roomRepo.Update(ctx, room)
	}

	if err := s.repo.Update(ctx, b); err != nil {
		return nil, fmt.Errorf("bookingService.Activate: %w", err)
	}
	s.log.Info("booking activated", zap.String("booking_id", id))
	return b, nil
}

// Complete owner closes an active booking (tenant moved out).
func (s *bookingService) Complete(ctx context.Context, id, ownerID string) (*domain.Booking, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("bookingService.Complete: %w", err)
	}
	b, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("bookingService.Complete: %w", err)
	}
	if b.OwnerID != ownerID {
		return nil, fmt.Errorf("bookingService.Complete: %w", domain.ErrUnauthorizedAccess)
	}
	if err := b.Complete(s.clock.Now()); err != nil {
		return nil, fmt.Errorf("bookingService.Complete: %w", apperrors.BadRequest(err.Error(), err))
	}

	// Free the room.
	room, roomErr := s.roomRepo.FindByID(ctx, b.RoomID)
	if roomErr == nil {
		room.SetStatus(domain.RoomStatusAvailable)
		room.Touch(s.clock.Now())
		_ = s.roomRepo.Update(ctx, room)
	}

	if err := s.repo.Update(ctx, b); err != nil {
		return nil, fmt.Errorf("bookingService.Complete: %w", err)
	}
	s.log.Info("booking completed", zap.String("booking_id", id))
	return b, nil
}

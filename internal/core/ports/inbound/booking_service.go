package inbound

import (
	"context"
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

//go:generate go run github.com/vektra/mockery/v2 --name=BookingService

// CreateBookingInput holds the data required to create a booking request.
type CreateBookingInput struct {
	RoomID    string
	TenantID  string
	StartDate time.Time
	EndDate   time.Time
	Notes     string
}

// BookingFilter defines query parameters for listing bookings.
type BookingFilter struct {
	OwnerID  string
	TenantID string
	RoomID   string
	Status   *domain.BookingStatus
	Page     int
	Limit    int
	SortDesc bool
}

// BookingService defines the inbound port for booking lifecycle management.
type BookingService interface {
	Create(ctx context.Context, in CreateBookingInput) (*domain.Booking, error)
	GetByID(ctx context.Context, id string) (*domain.Booking, error)
	List(ctx context.Context, f BookingFilter) ([]*domain.Booking, int64, error)
	Approve(ctx context.Context, id string, ownerID string) (*domain.Booking, error)
	Reject(ctx context.Context, id string, ownerID string, reason string) (*domain.Booking, error)
	Cancel(ctx context.Context, id string, requesterID string) (*domain.Booking, error)
	Activate(ctx context.Context, id string, ownerID string) (*domain.Booking, error)
	Complete(ctx context.Context, id string, ownerID string) (*domain.Booking, error)
}

package outbound

import (
	"context"
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
	"github.com/chatchomphu1000/go-starter/internal/core/ports/inbound"
)

//go:generate go run github.com/vektra/mockery/v2 --name=BookingRepository

// BookingRepository defines the outbound port for booking persistence.
type BookingRepository interface {
	Insert(ctx context.Context, b *domain.Booking) error
	FindByID(ctx context.Context, id string) (*domain.Booking, error)
	FindAll(ctx context.Context, f inbound.BookingFilter) ([]*domain.Booking, int64, error)
	Update(ctx context.Context, b *domain.Booking) error
	HasActiveBooking(ctx context.Context, roomID string, start, end time.Time) (bool, error)
	CountByStatus(ctx context.Context, ownerID string, status domain.BookingStatus) (int64, error)
}

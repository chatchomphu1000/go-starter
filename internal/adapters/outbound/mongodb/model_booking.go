package mongodb

import (
	"time"

	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

type bookingDoc struct {
	ID              string    `bson:"_id"`
	RoomID          string    `bson:"room_id"`
	TenantID        string    `bson:"tenant_id"`
	OwnerID         string    `bson:"owner_id"`
	Status          string    `bson:"status"`
	StartDate       time.Time `bson:"start_date"`
	EndDate         time.Time `bson:"end_date"`
	MonthlyRent     float64   `bson:"monthly_rent"`
	Deposit         float64   `bson:"deposit"`
	Notes           string    `bson:"notes"`
	RejectionReason string    `bson:"rejection_reason"`
	CreatedAt       time.Time `bson:"created_at"`
	UpdatedAt       time.Time `bson:"updated_at"`
}

func bookingFromDomain(b *domain.Booking) bookingDoc {
	return bookingDoc{
		ID:              b.ID,
		RoomID:          b.RoomID,
		TenantID:        b.TenantID,
		OwnerID:         b.OwnerID,
		Status:          string(b.Status),
		StartDate:       b.StartDate,
		EndDate:         b.EndDate,
		MonthlyRent:     b.MonthlyRent,
		Deposit:         b.Deposit,
		Notes:           b.Notes,
		RejectionReason: b.RejectionReason,
		CreatedAt:       b.CreatedAt,
		UpdatedAt:       b.UpdatedAt,
	}
}

func bookingToDomain(d bookingDoc) *domain.Booking {
	return &domain.Booking{
		ID:              d.ID,
		RoomID:          d.RoomID,
		TenantID:        d.TenantID,
		OwnerID:         d.OwnerID,
		Status:          domain.BookingStatus(d.Status),
		StartDate:       d.StartDate,
		EndDate:         d.EndDate,
		MonthlyRent:     d.MonthlyRent,
		Deposit:         d.Deposit,
		Notes:           d.Notes,
		RejectionReason: d.RejectionReason,
		CreatedAt:       d.CreatedAt,
		UpdatedAt:       d.UpdatedAt,
	}
}

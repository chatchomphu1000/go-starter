package domain

import (
	"fmt"
	"time"
)

// BookingStatus represents the lifecycle state of a booking.
type BookingStatus string

// Booking status constants.
const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusApproved  BookingStatus = "approved"
	BookingStatusRejected  BookingStatus = "rejected"
	BookingStatusActive    BookingStatus = "active"
	BookingStatusCompleted BookingStatus = "completed"
	BookingStatusCancelled BookingStatus = "cancelled"
)

// Booking represents a rental agreement request between a tenant and an owner.
type Booking struct {
	ID              string
	RoomID          string
	TenantID        string
	OwnerID         string
	Status          BookingStatus
	StartDate       time.Time
	EndDate         time.Time
	MonthlyRent     float64
	Deposit         float64
	Notes           string
	RejectionReason string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// NewBooking creates a validated Booking entity.
func NewBooking(
	id, roomID, tenantID, ownerID string,
	startDate, endDate time.Time,
	monthlyRent, deposit float64,
	notes string,
	now time.Time,
) (*Booking, error) {
	b := &Booking{
		ID:          id,
		RoomID:      roomID,
		TenantID:    tenantID,
		OwnerID:     ownerID,
		Status:      BookingStatusPending,
		StartDate:   startDate,
		EndDate:     endDate,
		MonthlyRent: monthlyRent,
		Deposit:     deposit,
		Notes:       notes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := b.Validate(); err != nil {
		return nil, err
	}
	return b, nil
}

// Validate enforces booking invariants.
func (b *Booking) Validate() error {
	if b.ID == "" {
		return fmt.Errorf("booking: id is required")
	}
	if b.RoomID == "" {
		return fmt.Errorf("booking: room_id is required")
	}
	if b.TenantID == "" {
		return fmt.Errorf("booking: tenant_id is required")
	}
	if b.OwnerID == "" {
		return fmt.Errorf("booking: owner_id is required")
	}
	if !b.EndDate.After(b.StartDate) {
		return fmt.Errorf("booking: end_date must be after start_date")
	}
	if b.MonthlyRent < 0 {
		return fmt.Errorf("booking: monthly_rent cannot be negative")
	}
	return nil
}

// Approve transitions the booking to approved state.
func (b *Booking) Approve(now time.Time) error {
	if b.Status != BookingStatusPending {
		return fmt.Errorf("%w: cannot approve a %s booking", ErrInvalidBookingTransition, b.Status)
	}
	b.Status = BookingStatusApproved
	b.UpdatedAt = now
	return nil
}

// Reject transitions the booking to rejected state.
func (b *Booking) Reject(reason string, now time.Time) error {
	if b.Status != BookingStatusPending {
		return fmt.Errorf("%w: cannot reject a %s booking", ErrInvalidBookingTransition, b.Status)
	}
	b.Status = BookingStatusRejected
	b.RejectionReason = reason
	b.UpdatedAt = now
	return nil
}

// Activate transitions an approved booking to active (tenant has moved in).
func (b *Booking) Activate(now time.Time) error {
	if b.Status != BookingStatusApproved {
		return fmt.Errorf("%w: cannot activate a %s booking", ErrInvalidBookingTransition, b.Status)
	}
	b.Status = BookingStatusActive
	b.UpdatedAt = now
	return nil
}

// Complete transitions the booking to completed.
func (b *Booking) Complete(now time.Time) error {
	if b.Status != BookingStatusActive {
		return fmt.Errorf("%w: cannot complete a %s booking", ErrInvalidBookingTransition, b.Status)
	}
	b.Status = BookingStatusCompleted
	b.UpdatedAt = now
	return nil
}

// Cancel transitions the booking to cancelled from pending or approved.
func (b *Booking) Cancel(now time.Time) error {
	if b.Status != BookingStatusPending && b.Status != BookingStatusApproved {
		return fmt.Errorf("%w: cannot cancel a %s booking", ErrInvalidBookingTransition, b.Status)
	}
	b.Status = BookingStatusCancelled
	b.UpdatedAt = now
	return nil
}

// Touch updates the UpdatedAt timestamp.
func (b *Booking) Touch(now time.Time) { b.UpdatedAt = now }

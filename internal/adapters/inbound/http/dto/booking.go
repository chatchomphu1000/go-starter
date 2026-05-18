package dto

import "github.com/chatchomphu1000/go-starter/internal/core/domain"

// CreateBookingRequest is the request body for creating a booking.
type CreateBookingRequest struct {
	RoomID    string `json:"room_id" validate:"required"`
	StartDate string `json:"start_date" validate:"required"` // YYYY-MM-DD
	EndDate   string `json:"end_date" validate:"required"`   // YYYY-MM-DD
	Notes     string `json:"notes" validate:"max=500"`
}

// RejectBookingRequest carries the rejection reason.
type RejectBookingRequest struct {
	Reason string `json:"reason" validate:"required,max=500"`
}

// BookingResponse is the API representation of a booking.
type BookingResponse struct {
	ID              string  `json:"id"`
	RoomID          string  `json:"room_id"`
	TenantID        string  `json:"tenant_id"`
	OwnerID         string  `json:"owner_id"`
	Status          string  `json:"status"`
	StartDate       string  `json:"start_date"`
	EndDate         string  `json:"end_date"`
	MonthlyRent     float64 `json:"monthly_rent"`
	Deposit         float64 `json:"deposit"`
	Notes           string  `json:"notes"`
	RejectionReason string  `json:"rejection_reason,omitempty"`
	CreatedAt       string  `json:"created_at"`
	UpdatedAt       string  `json:"updated_at"`
}

// BookingListResponse wraps a paginated list of bookings.
type BookingListResponse struct {
	Data  []BookingResponse `json:"data"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
}

// ToBookingResponse maps a domain Booking to BookingResponse.
func ToBookingResponse(b *domain.Booking) BookingResponse {
	return BookingResponse{
		ID:              b.ID,
		RoomID:          b.RoomID,
		TenantID:        b.TenantID,
		OwnerID:         b.OwnerID,
		Status:          string(b.Status),
		StartDate:       b.StartDate.Format("2006-01-02"),
		EndDate:         b.EndDate.Format("2006-01-02"),
		MonthlyRent:     b.MonthlyRent,
		Deposit:         b.Deposit,
		Notes:           b.Notes,
		RejectionReason: b.RejectionReason,
		CreatedAt:       b.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       b.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ToBookingListResponse builds a paginated booking list response.
func ToBookingListResponse(bookings []*domain.Booking, total int64, page, limit int) BookingListResponse {
	data := make([]BookingResponse, 0, len(bookings))
	for _, b := range bookings {
		data = append(data, ToBookingResponse(b))
	}
	return BookingListResponse{Data: data, Total: total, Page: page, Limit: limit}
}

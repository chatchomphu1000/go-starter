package domain

import (
	"fmt"
	"time"
)

// InvoiceStatus represents the lifecycle state of an invoice.
type InvoiceStatus string

// Invoice status constants.
const (
	InvoiceStatusDraft     InvoiceStatus = "draft"
	InvoiceStatusSent      InvoiceStatus = "sent"
	InvoiceStatusPaid      InvoiceStatus = "paid"
	InvoiceStatusOverdue   InvoiceStatus = "overdue"
	InvoiceStatusCancelled InvoiceStatus = "cancelled"
)

// InvoiceItem is a single line item on an invoice.
type InvoiceItem struct {
	Description string
	Quantity    int
	UnitPrice   float64
	Total       float64
}

// Invoice represents a billing document issued to a tenant.
type Invoice struct {
	ID          string
	BookingID   string
	TenantID    string
	OwnerID     string
	Items       []InvoiceItem
	TotalAmount float64
	Status      InvoiceStatus
	DueDate     time.Time
	PaidAt      *time.Time
	Month       int // billing month 1-12
	Year        int // billing year
	Notes       string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewInvoice creates a validated Invoice entity.
func NewInvoice(
	id, bookingID, tenantID, ownerID string,
	items []InvoiceItem,
	dueDate time.Time,
	month, year int,
	notes string,
	now time.Time,
) (*Invoice, error) {
	total := 0.0
	for _, item := range items {
		total += item.Total
	}

	inv := &Invoice{
		ID:          id,
		BookingID:   bookingID,
		TenantID:    tenantID,
		OwnerID:     ownerID,
		Items:       items,
		TotalAmount: total,
		Status:      InvoiceStatusDraft,
		DueDate:     dueDate,
		Month:       month,
		Year:        year,
		Notes:       notes,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := inv.Validate(); err != nil {
		return nil, err
	}
	return inv, nil
}

// Validate enforces invoice invariants.
func (inv *Invoice) Validate() error {
	if inv.ID == "" {
		return fmt.Errorf("invoice: id is required")
	}
	if inv.BookingID == "" {
		return fmt.Errorf("invoice: booking_id is required")
	}
	if inv.TotalAmount < 0 {
		return fmt.Errorf("invoice: total_amount cannot be negative")
	}
	if inv.Month < 1 || inv.Month > 12 {
		return fmt.Errorf("invoice: month must be between 1 and 12")
	}
	if inv.Year < 2020 {
		return fmt.Errorf("invoice: year is invalid")
	}
	return nil
}

// Send transitions the invoice from draft to sent.
func (inv *Invoice) Send(now time.Time) error {
	if inv.Status != InvoiceStatusDraft {
		return fmt.Errorf("invoice: only draft invoices can be sent")
	}
	inv.Status = InvoiceStatusSent
	inv.UpdatedAt = now
	return nil
}

// MarkPaid transitions the invoice to paid status.
func (inv *Invoice) MarkPaid(paidAt time.Time) error {
	if inv.Status == InvoiceStatusPaid {
		return fmt.Errorf("%w", ErrPaymentAlreadyPaid)
	}
	inv.Status = InvoiceStatusPaid
	inv.PaidAt = &paidAt
	inv.UpdatedAt = paidAt
	return nil
}

// MarkOverdue transitions the invoice to overdue.
func (inv *Invoice) MarkOverdue(now time.Time) {
	if inv.Status == InvoiceStatusSent {
		inv.Status = InvoiceStatusOverdue
		inv.UpdatedAt = now
	}
}

// Cancel voids the invoice.
func (inv *Invoice) Cancel(now time.Time) {
	inv.Status = InvoiceStatusCancelled
	inv.UpdatedAt = now
}

// Touch updates the UpdatedAt timestamp.
func (inv *Invoice) Touch(now time.Time) { inv.UpdatedAt = now }

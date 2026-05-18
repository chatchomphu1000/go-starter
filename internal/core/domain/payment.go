package domain

import (
	"fmt"
	"time"
)

// PaymentStatus represents the lifecycle state of a payment.
type PaymentStatus string

// Payment status constants.
const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusRefunded  PaymentStatus = "refunded"
)

// PaymentMethod is the mechanism used to complete the payment.
type PaymentMethod string

// Payment method constants.
const (
	PaymentMethodCreditCard   PaymentMethod = "credit_card"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
	PaymentMethodCash         PaymentMethod = "cash"
	PaymentMethodQRCode       PaymentMethod = "qr_code"
)

// Payment records a financial transaction for a booking or invoice.
type Payment struct {
	ID             string
	BookingID      string
	InvoiceID      string
	TenantID       string
	OwnerID        string
	Amount         float64
	Currency       string
	Status         PaymentStatus
	Method         PaymentMethod
	Gateway        string // stripe | omise | paypal | manual
	GatewayTxID    string
	WebhookPayload string // raw JSON from gateway webhook
	PaidAt         *time.Time
	Description    string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// NewPayment creates a validated Payment entity.
func NewPayment(
	id, bookingID, invoiceID, tenantID, ownerID string,
	amount float64,
	currency string,
	method PaymentMethod,
	gateway, description string,
	now time.Time,
) (*Payment, error) {
	p := &Payment{
		ID:          id,
		BookingID:   bookingID,
		InvoiceID:   invoiceID,
		TenantID:    tenantID,
		OwnerID:     ownerID,
		Amount:      amount,
		Currency:    currency,
		Status:      PaymentStatusPending,
		Method:      method,
		Gateway:     gateway,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

// Validate enforces payment invariants.
func (p *Payment) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("payment: id is required")
	}
	if p.Amount <= 0 {
		return fmt.Errorf("payment: amount must be positive")
	}
	return nil
}

// Complete marks the payment as successfully processed.
func (p *Payment) Complete(txID string, paidAt time.Time) error {
	if p.Status == PaymentStatusCompleted {
		return fmt.Errorf("%w", ErrPaymentAlreadyPaid)
	}
	p.Status = PaymentStatusCompleted
	p.GatewayTxID = txID
	p.PaidAt = &paidAt
	p.UpdatedAt = paidAt
	return nil
}

// Fail marks the payment as failed.
func (p *Payment) Fail(now time.Time) {
	p.Status = PaymentStatusFailed
	p.UpdatedAt = now
}

// Refund marks the payment as refunded.
func (p *Payment) Refund(now time.Time) error {
	if p.Status != PaymentStatusCompleted {
		return fmt.Errorf("payment: only completed payments can be refunded")
	}
	p.Status = PaymentStatusRefunded
	p.UpdatedAt = now
	return nil
}

// Touch updates the UpdatedAt timestamp.
func (p *Payment) Touch(now time.Time) { p.UpdatedAt = now }

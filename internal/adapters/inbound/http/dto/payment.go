package dto

import (
	"github.com/chatchomphu1000/go-starter/internal/core/domain"
)

// CreatePaymentRequest is the request body for creating a payment.
type CreatePaymentRequest struct {
	BookingID   string  `json:"booking_id" validate:"required"`
	InvoiceID   string  `json:"invoice_id"`
	Amount      float64 `json:"amount" validate:"required,gt=0"`
	Currency    string  `json:"currency" validate:"required,len=3"`
	Method      string  `json:"method" validate:"required,oneof=credit_card bank_transfer cash qr_code"`
	Gateway     string  `json:"gateway" validate:"oneof=stripe omise paypal manual"`
	Description string  `json:"description" validate:"max=500"`
}

// WebhookRequest is the payload from a payment gateway webhook.
type WebhookRequest struct {
	PaymentID   string `json:"payment_id" validate:"required"`
	GatewayTxID string `json:"gateway_tx_id"`
	Status      string `json:"status" validate:"required,oneof=completed failed refunded"`
}

// PaymentResponse is the API representation of a payment.
type PaymentResponse struct {
	ID          string  `json:"id"`
	BookingID   string  `json:"booking_id"`
	InvoiceID   string  `json:"invoice_id"`
	TenantID    string  `json:"tenant_id"`
	OwnerID     string  `json:"owner_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	Status      string  `json:"status"`
	Method      string  `json:"method"`
	Gateway     string  `json:"gateway"`
	GatewayTxID string  `json:"gateway_tx_id"`
	Description string  `json:"description"`
	PaidAt      *string `json:"paid_at,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// PaymentListResponse wraps a paginated list of payments.
type PaymentListResponse struct {
	Data  []PaymentResponse `json:"data"`
	Total int64             `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
}

// ToPaymentResponse maps a domain Payment to PaymentResponse.
func ToPaymentResponse(p *domain.Payment) PaymentResponse {
	r := PaymentResponse{
		ID:          p.ID,
		BookingID:   p.BookingID,
		InvoiceID:   p.InvoiceID,
		TenantID:    p.TenantID,
		OwnerID:     p.OwnerID,
		Amount:      p.Amount,
		Currency:    p.Currency,
		Status:      string(p.Status),
		Method:      string(p.Method),
		Gateway:     p.Gateway,
		GatewayTxID: p.GatewayTxID,
		Description: p.Description,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if p.PaidAt != nil {
		s := p.PaidAt.Format("2006-01-02T15:04:05Z")
		r.PaidAt = &s
	}
	return r
}

// ToPaymentListResponse builds a paginated payment list.
func ToPaymentListResponse(payments []*domain.Payment, total int64, page, limit int) PaymentListResponse {
	data := make([]PaymentResponse, 0, len(payments))
	for _, p := range payments {
		data = append(data, ToPaymentResponse(p))
	}
	return PaymentListResponse{Data: data, Total: total, Page: page, Limit: limit}
}
